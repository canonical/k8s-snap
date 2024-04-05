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
)

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
	if err := conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
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

	mu        sync.Mutex // guards the following fields
	remotes   []*remote
	pickCount int // for round robin
}

func (tp *tcpproxy) Run() error {
	tp.donec = make(chan struct{})
	if tp.MonitorInterval == 0 {
		tp.MonitorInterval = 5 * time.Minute
	}
	for _, srv := range tp.Endpoints {
		addr := fmt.Sprintf("%s:%d", srv.Target, srv.Port)
		tp.remotes = append(tp.remotes, &remote{srv: srv, addr: addr})
	}

	eps := []string{}
	for _, ep := range tp.Endpoints {
		eps = append(eps, fmt.Sprintf("%s:%d", ep.Target, ep.Port))
	}
	log.Printf("ready to proxy client requests to %v\n", eps)

	go tp.runMonitor()
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
		// TODO: add timeout
		out, err = net.Dial("tcp", remote.addr)
		if err == nil {
			break
		}
		remote.inactivate()
		log.Printf("deactivated endpoint %v for interval %v, error was %q", remote.addr, tp.MonitorInterval, err)
	}

	if out == nil {
		if err := in.Close(); err != nil {
			log.Printf("failed to close incoming connection: %v", err)
		}
		return
	}

	go func() {
		_, err := io.Copy(in, out)
		if err != nil {
			log.Printf("error copying from %v to %v: %v\n", out.RemoteAddr(), in.RemoteAddr(), err)
		}
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

func (tp *tcpproxy) Stop() {
	// graceful shutdown?
	// shutdown current connections?
	if err := tp.Listener.Close(); err != nil {
		log.Printf("failed to close listener: %v", err)
	}
	close(tp.donec)
}
