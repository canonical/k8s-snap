// This code is based on the tcpproxy implementation found in
// https://github.com/etcd-io/etcd/blob/v3.5.4/server/proxy/tcpproxy/userspace.go
//
// Original copyright notice follows:

// Copyright 2016 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package proxy

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/canonical/k8s/pkg/utils"
)

// Connection pool structures
type pooledConn struct {
	conn     net.Conn
	lastUsed time.Time
	inUse    bool
}

type connPool struct {
	mu          sync.Mutex
	connections map[string][]*pooledConn
	maxPerAddr  int
	maxIdleTime time.Duration
}

func newConnPool(maxPerAddr int, maxIdleTime time.Duration) *connPool {
	return &connPool{
		connections: make(map[string][]*pooledConn),
		maxPerAddr:  maxPerAddr,
		maxIdleTime: maxIdleTime,
	}
}

func (cp *connPool) get(addr string) (net.Conn, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Clean up expired connections first
	cp.cleanupExpired(addr)

	// Try to find an available connection
	if conns, exists := cp.connections[addr]; exists {
		for i, pc := range conns {
			if !pc.inUse {
				pc.inUse = true
				pc.lastUsed = time.Now()
				return &poolConnWrapper{pc, cp, addr, i}, nil
			}
		}
	}

	// No available connection, create a new one
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	pc := &pooledConn{
		conn:     conn,
		lastUsed: time.Now(),
		inUse:    true,
	}

	if cp.connections[addr] == nil {
		cp.connections[addr] = make([]*pooledConn, 0, cp.maxPerAddr)
	}

	if len(cp.connections[addr]) < cp.maxPerAddr {
		cp.connections[addr] = append(cp.connections[addr], pc)
		return &poolConnWrapper{pc, cp, addr, len(cp.connections[addr]) - 1}, nil
	}

	// Pool is full, return direct connection
	return conn, nil
}

func (cp *connPool) release(addr string, index int) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if conns, exists := cp.connections[addr]; exists && index < len(conns) {
		conns[index].inUse = false
		conns[index].lastUsed = time.Now()
	}
}

func (cp *connPool) cleanupExpired(addr string) {
	if conns, exists := cp.connections[addr]; exists {
		now := time.Now()
		for i := len(conns) - 1; i >= 0; i-- {
			pc := conns[i]
			if !pc.inUse && now.Sub(pc.lastUsed) > cp.maxIdleTime {
				pc.conn.Close()
				cp.connections[addr] = append(conns[:i], conns[i+1:]...)
			}
		}
	}
}

func (cp *connPool) close() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	for _, conns := range cp.connections {
		for _, pc := range conns {
			pc.conn.Close()
		}
	}
	cp.connections = make(map[string][]*pooledConn)
}

type poolConnWrapper struct {
	*pooledConn
	pool  *connPool
	addr  string
	index int
}

func (pcw *poolConnWrapper) Read(b []byte) (n int, err error) {
	return pcw.conn.Read(b)
}

func (pcw *poolConnWrapper) Write(b []byte) (n int, err error) {
	return pcw.conn.Write(b)
}

func (pcw *poolConnWrapper) Close() error {
	pcw.pool.release(pcw.addr, pcw.index)
	return nil // Don't actually close the underlying connection
}

func (pcw *poolConnWrapper) LocalAddr() net.Addr {
	return pcw.conn.LocalAddr()
}

func (pcw *poolConnWrapper) RemoteAddr() net.Addr {
	return pcw.conn.RemoteAddr()
}

func (pcw *poolConnWrapper) SetDeadline(t time.Time) error {
	return pcw.conn.SetDeadline(t)
}

func (pcw *poolConnWrapper) SetReadDeadline(t time.Time) error {
	return pcw.conn.SetReadDeadline(t)
}

func (pcw *poolConnWrapper) SetWriteDeadline(t time.Time) error {
	return pcw.conn.SetWriteDeadline(t)
}

type remote struct {
	mu       sync.Mutex
	srv      *net.SRV
	addr     string
	inactive bool
}

func (r *remote) inactivate() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inactive = true
}

func (r *remote) tryReactivate() error {
	conn, err := net.Dial("tcp", r.addr)
	if err != nil {
		return err
	}
	conn.Close()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inactive = false
	return nil
}

func (r *remote) isActive() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return !r.inactive
}

type tcpproxy struct {
	Listener        net.Listener
	Endpoints       []*net.SRV
	MonitorInterval time.Duration

	donec chan struct{}
	pool  *connPool

	mu        sync.Mutex // guards the following fields
	remotes   []*remote
	pickCount int // for round robin
}

func (tp *tcpproxy) Run() error {
	tp.donec = make(chan struct{})
	tp.pool = newConnPool(10, 5*time.Minute) // 10 conns per addr, 5min idle timeout

	if tp.MonitorInterval == 0 {
		tp.MonitorInterval = 5 * time.Minute
	}
	for _, srv := range tp.Endpoints {
		ip := net.ParseIP(srv.Target)
		addr := fmt.Sprintf("%s:%d", utils.ToIPString(ip), srv.Port)
		tp.remotes = append(tp.remotes, &remote{srv: srv, addr: addr})
	}

	eps := []string{}
	for _, ep := range tp.Endpoints {
		eps = append(eps, fmt.Sprintf("%s:%d", ep.Target, ep.Port))
	}
	log.Printf("ready to proxy client requests to %v\n", eps)

	go tp.runMonitor()
	go tp.runPoolCleaner()
	for {
		in, err := tp.Listener.Accept()
		if err != nil {
			return err
		}

		go tp.serve(in)
	}
}

func (tp *tcpproxy) pick() *remote {
	var weighted []*remote
	var unweighted []*remote

	bestPr := uint16(65535)
	w := 0
	// find best priority class
	for _, r := range tp.remotes {
		switch {
		case !r.isActive():
		case r.srv.Priority < bestPr:
			bestPr = r.srv.Priority
			w = 0
			weighted = nil
			unweighted = nil
			fallthrough
		case r.srv.Priority == bestPr:
			if r.srv.Weight > 0 {
				weighted = append(weighted, r)
				w += int(r.srv.Weight)
			} else {
				unweighted = append(unweighted, r)
			}
		}
	}
	if weighted != nil {
		if len(unweighted) > 0 && rand.Intn(100) == 1 {
			// In the presence of records containing weights greater
			// than 0, records with weight 0 should have a very small
			// chance of being selected.
			r := unweighted[tp.pickCount%len(unweighted)]
			tp.pickCount++
			return r
		}
		// choose a uniform random number between 0 and the sum computed
		// (inclusive), and select the RR whose running sum value is the
		// first in the selected order
		choose := rand.Intn(w)
		for i := 0; i < len(weighted); i++ {
			choose -= int(weighted[i].srv.Weight)
			if choose <= 0 {
				return weighted[i]
			}
		}
	}
	if unweighted != nil {
		for i := 0; i < len(tp.remotes); i++ {
			picked := tp.remotes[tp.pickCount%len(tp.remotes)]
			tp.pickCount++
			if picked.isActive() {
				return picked
			}
		}
	}
	return nil
}

func (tp *tcpproxy) serve(in net.Conn) {
	var (
		err error
		out net.Conn
	)

	for {
		tp.mu.Lock()
		remote := tp.pick()
		tp.mu.Unlock()
		if remote == nil {
			break
		}
		// Use connection pool instead of direct dial
		out, err = tp.pool.get(remote.addr)
		if err == nil {
			break
		}
		remote.inactivate()
		log.Printf("deactivated endpoint %v for interval %v, error was %q", remote.addr, tp.MonitorInterval, err)
	}

	if out == nil {
		in.Close()
		return
	}

	go func() {
		io.Copy(in, out)
		in.Close()
		out.Close()
	}()

	io.Copy(out, in)
	out.Close()
	in.Close()
}

func (tp *tcpproxy) runMonitor() {
	for {
		select {
		case <-time.After(tp.MonitorInterval):
			tp.mu.Lock()
			for _, rem := range tp.remotes {
				if rem.isActive() {
					continue
				}
				go func(r *remote) {
					if err := r.tryReactivate(); err != nil {
						log.Printf("failed to activate endpoint %v (stay inactive for another interval %v)\n", r.addr, tp.MonitorInterval)
					} else {
						log.Printf("activated endpoint %v\n", r.addr)
					}
				}(rem)
			}
			tp.mu.Unlock()
		case <-tp.donec:
			return
		}
	}
}

func (tp *tcpproxy) runPoolCleaner() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tp.pool.mu.Lock()
			for addr := range tp.pool.connections {
				tp.pool.cleanupExpired(addr)
			}
			tp.pool.mu.Unlock()
		case <-tp.donec:
			return
		}
	}
}

func (tp *tcpproxy) Stop() {
	// graceful shutdown?
	// shutdown current connections?
	tp.Listener.Close()
	if tp.pool != nil {
		tp.pool.close()
	}
	close(tp.donec)
}
