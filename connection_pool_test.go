package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// Mock system for testing
type MockSystem struct {
	DSConnectionPool chan net.Conn
}

var testSystem = &MockSystem{
	DSConnectionPool: make(chan net.Conn, 4), // Pool size of 4
}

// Mock connection that implements net.Conn interface
type MockConn struct {
	id     int
	closed bool
	mu     sync.Mutex
}

func (m *MockConn) Read(b []byte) (n int, err error)  { return 0, nil }
func (m *MockConn) Write(b []byte) (n int, err error) { return len(b), nil }
func (m *MockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	fmt.Printf("MockConn %d closed\n", m.id)
	return nil
}
func (m *MockConn) LocalAddr() net.Addr                { return nil }
func (m *MockConn) RemoteAddr() net.Addr               { return nil }
func (m *MockConn) SetDeadline(t time.Time) error      { return nil }
func (m *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *MockConn) SetWriteDeadline(t time.Time) error { return nil }

func (m *MockConn) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

// Mock version of the middleware for testing
func MockConnectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Current connection pool size: %d\n", len(testSystem.DSConnectionPool))

		var conn net.Conn
		select {
		case conn = <-testSystem.DSConnectionPool:
			fmt.Println("Got connection from pool")
		default:
			fmt.Println("No connections available, refusing request")
			http.Error(w, "No connections available", http.StatusServiceUnavailable)
			return
		}

		// Pass connection in context
		ctx := context.WithValue(r.Context(), "conn", conn)

		// Return connection to pool after request
		defer func() {
			testSystem.DSConnectionPool <- conn
			fmt.Println("Returned connection to pool")
		}()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Initialize test connection pool
func initTestConnectionPool(poolSize int) {
	for i := 0; i < poolSize; i++ {
		conn := &MockConn{id: i + 1}
		testSystem.DSConnectionPool <- conn
		fmt.Printf("Added mock connection %d to pool\n", i+1)
	}
}

// Test handler that simulates work
func testHandler(w http.ResponseWriter, r *http.Request) {
	conn := r.Context().Value("conn")
	if conn == nil {
		http.Error(w, "No connection in context", http.StatusInternalServerError)
		return
	}

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success"))
}

func TestConnectionPoolBehavior(t *testing.T) {
	// Initialize the pool with 4 connections
	initTestConnectionPool(4)

	// Create test server with middleware
	handler := MockConnectionMiddleware(http.HandlerFunc(testHandler))
	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("Pool should handle exactly 4 concurrent requests", func(t *testing.T) {
		var wg sync.WaitGroup
		results := make(chan int, 10)

		// Start 4 concurrent long-running requests
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func(requestID int) {
				defer wg.Done()

				client := &http.Client{Timeout: 5 * time.Second}
				resp, err := client.Get(server.URL)
				if err != nil {
					t.Errorf("Request %d failed: %v", requestID, err)
					results <- 500
					return
				}
				defer resp.Body.Close()

				fmt.Printf("Request %d completed with status: %d\n", requestID, resp.StatusCode)
				results <- resp.StatusCode
			}(i + 1)
		}

		wg.Wait()
		close(results)

		// Check that all 4 requests succeeded
		successCount := 0
		for statusCode := range results {
			if statusCode == 200 {
				successCount++
			}
		}

		if successCount != 4 {
			t.Errorf("Expected 4 successful requests, got %d", successCount)
		}
	})

	t.Run("Pool should refuse 5th concurrent request", func(t *testing.T) {
		var wg sync.WaitGroup
		results := make(chan int, 10)

		// Start 5 concurrent requests (4 should succeed, 1 should fail)
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(requestID int) {
				defer wg.Done()

				client := &http.Client{Timeout: 5 * time.Second}
				resp, err := client.Get(server.URL)
				if err != nil {
					t.Errorf("Request %d failed: %v", requestID, err)
					results <- 500
					return
				}
				defer resp.Body.Close()

				fmt.Printf("Request %d completed with status: %d\n", requestID, resp.StatusCode)
				results <- resp.StatusCode
			}(i + 1)
		}

		wg.Wait()
		close(results)

		// Count successful vs failed requests
		successCount := 0
		serviceUnavailableCount := 0

		for statusCode := range results {
			switch statusCode {
			case 200:
				successCount++
			case 503: // Service Unavailable
				serviceUnavailableCount++
			}
		}

		fmt.Printf("Success: %d, Service Unavailable: %d\n", successCount, serviceUnavailableCount)

		if successCount != 4 {
			t.Errorf("Expected 4 successful requests, got %d", successCount)
		}
		if serviceUnavailableCount != 1 {
			t.Errorf("Expected 1 service unavailable response, got %d", serviceUnavailableCount)
		}
	})

	t.Run("Pool should recover after requests complete", func(t *testing.T) {
		// Wait a bit for any ongoing requests to complete
		time.Sleep(200 * time.Millisecond)

		// Pool should be full again
		poolSize := len(testSystem.DSConnectionPool)
		if poolSize != 4 {
			t.Errorf("Expected pool size to be 4 after requests complete, got %d", poolSize)
		}

		// Should be able to make new requests
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(server.URL)
		if err != nil {
			t.Errorf("Recovery request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected 200 status code after recovery, got %d", resp.StatusCode)
		}
	})
}

// Benchmark to test performance under load
func BenchmarkConnectionPool(b *testing.B) {
	// Reset pool for benchmark
	testSystem.DSConnectionPool = make(chan net.Conn, 4)
	initTestConnectionPool(4)

	handler := MockConnectionMiddleware(http.HandlerFunc(testHandler))
	server := httptest.NewServer(handler)
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(server.URL)
			if err != nil {
				b.Errorf("Benchmark request failed: %v", err)
				continue
			}
			resp.Body.Close()
		}
	})
}

// Test to verify connection reuse
func TestConnectionReuse(t *testing.T) {
	// Reset pool
	testSystem.DSConnectionPool = make(chan net.Conn, 4)
	initTestConnectionPool(4)

	handler := MockConnectionMiddleware(http.HandlerFunc(testHandler))
	server := httptest.NewServer(handler)
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	// Make several sequential requests
	for i := 0; i < 10; i++ {
		resp, err := client.Get(server.URL)
		if err != nil {
			t.Errorf("Request %d failed: %v", i+1, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Request %d returned status %d, expected 200", i+1, resp.StatusCode)
		}
	}

	// Pool should still have 4 connections
	poolSize := len(testSystem.DSConnectionPool)
	if poolSize != 4 {
		t.Errorf("Expected pool size to remain 4 after sequential requests, got %d", poolSize)
	}
}
