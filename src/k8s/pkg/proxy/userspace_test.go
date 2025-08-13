package proxy

import (
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

// Mock server for testing
type mockServer struct {
	listener  net.Listener
	addr      string
	connCount int
	mu        sync.Mutex
}

func newMockServer() (*mockServer, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	ms := &mockServer{
		listener: listener,
		addr:     listener.Addr().String(),
	}

	go ms.run()
	return ms, nil
}

func (ms *mockServer) run() {
	for {
		conn, err := ms.listener.Accept()
		if err != nil {
			return
		}

		ms.mu.Lock()
		ms.connCount++
		ms.mu.Unlock()

		go func(c net.Conn) {
			defer c.Close()
			// Echo server - read and write back
			io.Copy(c, c)
		}(conn)
	}
}

func (ms *mockServer) getConnCount() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.connCount
}

func (ms *mockServer) close() {
	ms.listener.Close()
}

func TestConnPool_Basic(t *testing.T) {
	pool := newConnPool(5, 30*time.Second)
	defer pool.close()

	// Create mock server
	server, err := newMockServer()
	if err != nil {
		t.Fatal(err)
	}
	defer server.close()

	// Test getting a connection
	conn1, err := pool.get(server.addr)
	if err != nil {
		t.Fatal(err)
	}

	// Test writing and reading
	testData := "hello world"
	_, err = conn1.Write([]byte(testData))
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, len(testData))
	_, err = conn1.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	if string(buf) != testData {
		t.Errorf("Expected %s, got %s", testData, string(buf))
	}

	// Close connection (should return to pool)
	conn1.Close()

	// Get another connection - should reuse the pooled one
	conn2, err := pool.get(server.addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close()

	// Verify only one connection was created to the server
	time.Sleep(100 * time.Millisecond) // Give server time to register connections
	if count := server.getConnCount(); count != 1 {
		t.Errorf("Expected 1 connection to server, got %d", count)
	}
}

func TestConnPool_MaxConnections(t *testing.T) {
	maxConns := 3
	pool := newConnPool(maxConns, 30*time.Second)
	defer pool.close()

	server, err := newMockServer()
	if err != nil {
		t.Fatal(err)
	}
	defer server.close()

	var conns []net.Conn

	// Get max connections
	for i := 0; i < maxConns; i++ {
		conn, err := pool.get(server.addr)
		if err != nil {
			t.Fatal(err)
		}
		conns = append(conns, conn)
	}

	// Next connection should be direct (not pooled)
	directConn, err := pool.get(server.addr)
	if err != nil {
		t.Fatal(err)
	}
	defer directConn.Close()

	// Close all pooled connections
	for _, conn := range conns {
		conn.Close()
	}

	// Verify we have maxConns + 1 connections (maxConns pooled + 1 direct)
	time.Sleep(100 * time.Millisecond)
	if count := server.getConnCount(); count != maxConns+1 {
		t.Errorf("Expected %d connections to server, got %d", maxConns+1, count)
	}
}

func TestConnPool_IdleTimeout(t *testing.T) {
	idleTimeout := 100 * time.Millisecond
	pool := newConnPool(5, idleTimeout)
	defer pool.close()

	server, err := newMockServer()
	if err != nil {
		t.Fatal(err)
	}
	defer server.close()

	// Get and close a connection
	conn, err := pool.get(server.addr)
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()

	// Wait for idle timeout + some buffer
	time.Sleep(idleTimeout + 50*time.Millisecond)

	// Manually trigger cleanup
	pool.mu.Lock()
	pool.cleanupExpired(server.addr)
	poolSize := len(pool.connections[server.addr])
	pool.mu.Unlock()

	if poolSize != 0 {
		t.Errorf("Expected pool to be empty after idle timeout, got %d connections", poolSize)
	}
}

func TestTCPProxy_WithoutPool(t *testing.T) {
	// Create a version of tcpproxy that doesn't use connection pooling
	// for comparison in benchmarks
	server, err := newMockServer()
	if err != nil {
		t.Fatal(err)
	}
	defer server.close()

	// Test direct connections without pooling
	for i := 0; i < 10; i++ {
		conn, err := net.Dial("tcp", server.addr)
		if err != nil {
			t.Fatal(err)
		}

		testData := fmt.Sprintf("test%d", i)
		conn.Write([]byte(testData))

		buf := make([]byte, len(testData))
		conn.Read(buf)
		conn.Close()

		if string(buf) != testData {
			t.Errorf("Expected %s, got %s", testData, string(buf))
		}
	}

	// Should have 10 connections
	time.Sleep(100 * time.Millisecond)
	if count := server.getConnCount(); count != 10 {
		t.Errorf("Expected 10 connections without pooling, got %d", count)
	}
}

// Benchmark without connection pooling
func BenchmarkWithoutConnectionPool(b *testing.B) {
	server, err := newMockServer()
	if err != nil {
		b.Fatal(err)
	}
	defer server.close()

	testData := []byte("benchmark test data")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Create new connection for each request (no pooling)
			conn, err := net.Dial("tcp", server.addr)
			if err != nil {
				b.Fatal(err)
			}

			// Write test data
			_, err = conn.Write(testData)
			if err != nil {
				b.Fatal(err)
			}

			// Read response
			buf := make([]byte, len(testData))
			_, err = conn.Read(buf)
			if err != nil {
				b.Fatal(err)
			}

			conn.Close()
		}
	})
}

// Benchmark with connection pooling
func BenchmarkWithConnectionPool(b *testing.B) {
	pool := newConnPool(10, 5*time.Minute)
	defer pool.close()

	server, err := newMockServer()
	if err != nil {
		b.Fatal(err)
	}
	defer server.close()

	testData := []byte("benchmark test data")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Get connection from pool
			conn, err := pool.get(server.addr)
			if err != nil {
				b.Fatal(err)
			}

			// Write test data
			_, err = conn.Write(testData)
			if err != nil {
				b.Fatal(err)
			}

			// Read response
			buf := make([]byte, len(testData))
			_, err = conn.Read(buf)
			if err != nil {
				b.Fatal(err)
			}

			// Return to pool
			conn.Close()
		}
	})
}

// Benchmark connection establishment overhead
func BenchmarkConnectionEstablishment(b *testing.B) {
	server, err := newMockServer()
	if err != nil {
		b.Fatal(err)
	}
	defer server.close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn, err := net.Dial("tcp", server.addr)
		if err != nil {
			b.Fatal(err)
		}
		conn.Close()
	}
}

// Benchmark pooled connection reuse
func BenchmarkPooledConnectionReuse(b *testing.B) {
	pool := newConnPool(10, 5*time.Minute)
	defer pool.close()

	server, err := newMockServer()
	if err != nil {
		b.Fatal(err)
	}
	defer server.close()

	// Pre-populate pool with one connection
	conn, err := pool.get(server.addr)
	if err != nil {
		b.Fatal(err)
	}
	conn.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn, err := pool.get(server.addr)
		if err != nil {
			b.Fatal(err)
		}
		conn.Close()
	}
}

// Test concurrent access to connection pool
func TestConnPool_Concurrent(t *testing.T) {
	pool := newConnPool(5, 30*time.Second)
	defer pool.close()

	server, err := newMockServer()
	if err != nil {
		t.Fatal(err)
	}
	defer server.close()

	var wg sync.WaitGroup
	concurrency := 20
	requestsPerGoroutine := 10

	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				conn, err := pool.get(server.addr)
				if err != nil {
					t.Errorf("Goroutine %d, request %d: %v", goroutineID, j, err)
					return
				}

				testData := fmt.Sprintf("goroutine%d-request%d", goroutineID, j)
				conn.Write([]byte(testData))

				buf := make([]byte, len(testData))
				conn.Read(buf)
				conn.Close()

				if string(buf) != testData {
					t.Errorf("Expected %s, got %s", testData, string(buf))
				}
			}
		}(i)
	}

	wg.Wait()

	// Check that we didn't create too many connections
	time.Sleep(100 * time.Millisecond)
	connCount := server.getConnCount()

	// We should have created fewer connections than total requests due to pooling
	totalRequests := concurrency * requestsPerGoroutine
	if connCount >= totalRequests {
		t.Errorf("Connection pooling not effective: created %d connections for %d requests", connCount, totalRequests)
	}

	t.Logf("Created %d connections for %d concurrent requests (%.1f%% efficiency)",
		connCount, totalRequests, float64(totalRequests-connCount)/float64(totalRequests)*100)
}

// Benchmark concurrent access with and without pooling
func BenchmarkConcurrentWithoutPool(b *testing.B) {
	server, err := newMockServer()
	if err != nil {
		b.Fatal(err)
	}
	defer server.close()

	testData := []byte("concurrent test")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			conn, err := net.Dial("tcp", server.addr)
			if err != nil {
				b.Fatal(err)
			}

			conn.Write(testData)
			buf := make([]byte, len(testData))
			conn.Read(buf)
			conn.Close()
		}
	})
}

func BenchmarkConcurrentWithPool(b *testing.B) {
	pool := newConnPool(10, 5*time.Minute)
	defer pool.close()

	server, err := newMockServer()
	if err != nil {
		b.Fatal(err)
	}
	defer server.close()

	testData := []byte("concurrent test")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			conn, err := pool.get(server.addr)
			if err != nil {
				b.Fatal(err)
			}

			conn.Write(testData)
			buf := make([]byte, len(testData))
			conn.Read(buf)
			conn.Close()
		}
	})
}

// Test pool cleanup functionality
func TestConnPool_Cleanup(t *testing.T) {
	pool := newConnPool(3, 50*time.Millisecond)
	defer pool.close()

	server, err := newMockServer()
	if err != nil {
		t.Fatal(err)
	}
	defer server.close()

	// Create and close some connections
	for range 3 {
		conn, err := pool.get(server.addr)
		if err != nil {
			t.Fatal(err)
		}
		conn.Close()
	}

	// Wait for idle timeout
	time.Sleep(100 * time.Millisecond)

	// Trigger cleanup
	pool.mu.Lock()
	pool.cleanupExpired(server.addr)
	remaining := len(pool.connections[server.addr])
	pool.mu.Unlock()

	if remaining != 0 {
		t.Errorf("Expected 0 connections after cleanup, got %d", remaining)
	}
}
