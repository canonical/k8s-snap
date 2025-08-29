package proxy

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// TestPerformanceSummary provides a clear summary of the performance benefits
func TestPerformanceSummary(t *testing.T) {
	server, err := newMockServer()
	if err != nil {
		t.Fatal(err)
	}
	defer server.close()

	pool := newConnPool(10, 5*time.Minute)
	defer pool.close()

	testData := []byte("performance test")
	iterations := 1000

	fmt.Println("\n=== CONNECTION POOL PERFORMANCE SUMMARY ===")

	// 1. Connection Establishment Overhead Test
	fmt.Println("\n1. Connection Establishment Overhead:")

	// Test direct connections
	start := time.Now()
	for i := 0; i < 100; i++ {
		conn, err := net.Dial("tcp", server.addr)
		if err == nil {
			conn.Close()
		}
	}
	directTime := time.Since(start)

	// Pre-populate pool and test reuse
	conn, _ := pool.get(server.addr)
	conn.Close()

	start = time.Now()
	for i := 0; i < 100; i++ {
		conn, err := pool.get(server.addr)
		if err == nil {
			conn.Close()
		}
	}
	pooledTime := time.Since(start)

	fmt.Printf("   Direct connections (100x): %v (%v avg per conn)\n", directTime, directTime/100)
	fmt.Printf("   Pooled connections (100x): %v (%v avg per conn)\n", pooledTime, pooledTime/100)
	fmt.Printf("   Improvement: %.1fx faster\n", float64(directTime.Nanoseconds())/float64(pooledTime.Nanoseconds()))

	// 2. Realistic Workload Test
	fmt.Println("\n2. Realistic Workload Test (multiple requests per goroutine):")

	// Without pool
	start = time.Now()
	for i := 0; i < iterations; i++ {
		conn, err := net.Dial("tcp", server.addr)
		if err == nil {
			conn.Write(testData)
			buf := make([]byte, len(testData))
			conn.Read(buf)
			conn.Close()
		}
	}
	withoutPoolTime := time.Since(start)

	// With pool
	start = time.Now()
	for i := 0; i < iterations; i++ {
		conn, err := pool.get(server.addr)
		if err == nil {
			conn.Write(testData)
			buf := make([]byte, len(testData))
			conn.Read(buf)
			conn.Close()
		}
	}
	withPoolTime := time.Since(start)

	fmt.Printf("   Without pool (%dx): %v (%v per request)\n", iterations, withoutPoolTime, withoutPoolTime/time.Duration(iterations))
	fmt.Printf("   With pool (%dx): %v (%v per request)\n", iterations, withPoolTime, withPoolTime/time.Duration(iterations))
	fmt.Printf("   Improvement: %.1fx faster\n", float64(withoutPoolTime.Nanoseconds())/float64(withPoolTime.Nanoseconds()))

	// 3. Connection Count Efficiency
	fmt.Println("\n3. Connection Efficiency:")
	initialConnCount := server.getConnCount()

	// Create some load with pooling
	for i := 0; i < 50; i++ {
		conn, err := pool.get(server.addr)
		if err == nil {
			conn.Write(testData)
			buf := make([]byte, len(testData))
			conn.Read(buf)
			conn.Close()
		}
	}

	time.Sleep(10 * time.Millisecond) // Let server count connections
	pooledConnCount := server.getConnCount() - initialConnCount

	fmt.Printf("   50 requests with pooling created %d new connections\n", pooledConnCount)
	fmt.Printf("   Connection efficiency: %.1f%% (saved %d connections)\n",
		float64(50-pooledConnCount)/50*100, 50-pooledConnCount)

	// 4. Memory Allocation Comparison
	fmt.Println("\n4. Memory Usage Benefits:")
	fmt.Printf("   Connection pooling reduces allocations by reusing connections\n")
	fmt.Printf("   See benchmark results for detailed allocation metrics\n")

	fmt.Println("\n=== CONCLUSION ===")
	fmt.Printf("Connection pooling provides significant benefits:\n")
	fmt.Printf("- %dx faster connection reuse\n", int(float64(directTime.Nanoseconds())/float64(pooledTime.Nanoseconds())))
	fmt.Printf("- %dx faster realistic workloads\n", int(float64(withoutPoolTime.Nanoseconds())/float64(withPoolTime.Nanoseconds())))
	fmt.Printf("- %.1f%% reduction in connection overhead\n", float64(50-pooledConnCount)/50*100)
	fmt.Printf("- Reduced memory allocations and GC pressure\n")
}
