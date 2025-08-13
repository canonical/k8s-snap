package proxy

import (
	"net"
	"sync"
	"testing"
	"time"
)

// More realistic benchmark that simulates multiple requests to the same endpoint
// over time (which is where connection pooling shines)
func BenchmarkRealisticWorkload_WithoutPool(b *testing.B) {
	server, err := newMockServer()
	if err != nil {
		b.Fatal(err)
	}
	defer server.close()

	testData := []byte("realistic workload test")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Simulate multiple requests per connection lifecycle
			for i := 0; i < 5; i++ {
				conn, err := net.Dial("tcp", server.addr)
				if err != nil {
					b.Fatal(err)
				}

				conn.Write(testData)
				buf := make([]byte, len(testData))
				conn.Read(buf)
				conn.Close()
			}
		}
	})
}

func BenchmarkRealisticWorkload_WithPool(b *testing.B) {
	pool := newConnPool(20, 5*time.Minute) // Larger pool for concurrent access
	defer pool.close()

	server, err := newMockServer()
	if err != nil {
		b.Fatal(err)
	}
	defer server.close()

	testData := []byte("realistic workload test")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Simulate multiple requests per connection lifecycle
			for i := 0; i < 5; i++ {
				conn, err := pool.get(server.addr)
				if err != nil {
					b.Fatal(err)
				}

				conn.Write(testData)
				buf := make([]byte, len(testData))
				conn.Read(buf)
				conn.Close() // Returns to pool
			}
		}
	})
}

// Test sustained load over time
func BenchmarkSustainedLoad_WithoutPool(b *testing.B) {
	server, err := newMockServer()
	if err != nil {
		b.Fatal(err)
	}
	defer server.close()

	testData := []byte("sustained load test")

	b.ResetTimer()

	// Run for a specific duration to measure sustained throughput
	start := time.Now()
	operations := 0

	for time.Since(start) < time.Second {
		for i := 0; i < 100; i++ { // Batch operations
			conn, err := net.Dial("tcp", server.addr)
			if err != nil {
				b.Fatal(err)
			}

			conn.Write(testData)
			buf := make([]byte, len(testData))
			conn.Read(buf)
			conn.Close()
			operations++
		}
	}

	b.ReportMetric(float64(operations), "ops/sec")
}

func BenchmarkSustainedLoad_WithPool(b *testing.B) {
	pool := newConnPool(50, 5*time.Minute)
	defer pool.close()

	server, err := newMockServer()
	if err != nil {
		b.Fatal(err)
	}
	defer server.close()

	testData := []byte("sustained load test")

	b.ResetTimer()

	// Run for a specific duration to measure sustained throughput
	start := time.Now()
	operations := 0

	for time.Since(start) < time.Second {
		for i := 0; i < 100; i++ { // Batch operations
			conn, err := pool.get(server.addr)
			if err != nil {
				b.Fatal(err)
			}

			conn.Write(testData)
			buf := make([]byte, len(testData))
			conn.Read(buf)
			conn.Close()
			operations++
		}
	}

	b.ReportMetric(float64(operations), "ops/sec")
}

// Benchmark that measures latency under load
func BenchmarkLatencyUnderLoad_WithoutPool(b *testing.B) {
	server, err := newMockServer()
	if err != nil {
		b.Fatal(err)
	}
	defer server.close()

	testData := []byte("latency test")

	// Create background load
	var wg sync.WaitGroup
	stopLoad := make(chan struct{})

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopLoad:
					return
				default:
					conn, err := net.Dial("tcp", server.addr)
					if err == nil {
						conn.Write(testData)
						buf := make([]byte, len(testData))
						conn.Read(buf)
						conn.Close()
					}
				}
			}
		}()
	}

	b.ResetTimer()

	// Measure latency of individual operations under load
	for i := 0; i < b.N; i++ {
		start := time.Now()
		conn, err := net.Dial("tcp", server.addr)
		if err != nil {
			b.Fatal(err)
		}

		conn.Write(testData)
		buf := make([]byte, len(testData))
		conn.Read(buf)
		conn.Close()

		elapsed := time.Since(start)
		b.ReportMetric(float64(elapsed.Nanoseconds()), "ns/op")
	}

	close(stopLoad)
	wg.Wait()
}

func BenchmarkLatencyUnderLoad_WithPool(b *testing.B) {
	pool := newConnPool(20, 5*time.Minute)
	defer pool.close()

	server, err := newMockServer()
	if err != nil {
		b.Fatal(err)
	}
	defer server.close()

	testData := []byte("latency test")

	// Create background load using the pool
	var wg sync.WaitGroup
	stopLoad := make(chan struct{})

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopLoad:
					return
				default:
					conn, err := pool.get(server.addr)
					if err == nil {
						conn.Write(testData)
						buf := make([]byte, len(testData))
						conn.Read(buf)
						conn.Close()
					}
				}
			}
		}()
	}

	b.ResetTimer()

	// Measure latency of individual operations under load
	for i := 0; i < b.N; i++ {
		start := time.Now()
		conn, err := pool.get(server.addr)
		if err != nil {
			b.Fatal(err)
		}

		conn.Write(testData)
		buf := make([]byte, len(testData))
		conn.Read(buf)
		conn.Close()

		elapsed := time.Since(start)
		b.ReportMetric(float64(elapsed.Nanoseconds()), "ns/op")
	}

	close(stopLoad)
	wg.Wait()
}

// Test connection establishment overhead specifically
func TestConnectionEstablishmentOverhead(t *testing.T) {
	server, err := newMockServer()
	if err != nil {
		t.Fatal(err)
	}
	defer server.close()

	pool := newConnPool(10, 5*time.Minute)
	defer pool.close()

	// Measure time for direct connection
	start := time.Now()
	directConn, err := net.Dial("tcp", server.addr)
	if err != nil {
		t.Fatal(err)
	}
	directTime := time.Since(start)
	directConn.Close()

	// Pre-populate pool
	poolConn, err := pool.get(server.addr)
	if err != nil {
		t.Fatal(err)
	}
	poolConn.Close()

	// Measure time for pooled connection reuse
	start = time.Now()
	pooledConn, err := pool.get(server.addr)
	if err != nil {
		t.Fatal(err)
	}
	pooledTime := time.Since(start)
	pooledConn.Close()

	t.Logf("Direct connection establishment: %v", directTime)
	t.Logf("Pooled connection reuse: %v", pooledTime)
	t.Logf("Improvement factor: %.2fx faster", float64(directTime.Nanoseconds())/float64(pooledTime.Nanoseconds()))

	if pooledTime >= directTime {
		t.Logf("Warning: Pooled connection not faster than direct connection")
	}
}
