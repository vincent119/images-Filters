package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunBenchmark(t *testing.T) {
	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Millisecond) // Simulate some work
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	// Short duration for testing
	duration := 100 * time.Millisecond
	concurrency := 2

	totalDur, success, errors, results := runBenchmark(ts.URL, concurrency, 0, duration, nil)

	assert.True(t, totalDur >= duration)
	assert.True(t, success > 0)
	assert.Equal(t, int64(0), errors)

	// Check results channel
	count := 0
	for range results {
		count++
	}
	assert.Equal(t, int(success), count)

	// Test reporting
	var params bytes.Buffer
	// Re-open channel for printReport if needed?
	// The runBenchmark closed it, so we can't iterate it again in printReport if we drained it.
	// Actually printReport iterates the channel.
	// But in my test above I drained it.
	// I should test printReport separately or not drain it above.

	// Let's test printReport with fresh data
	results2 := make(chan time.Duration, 10)
	results2 <- 10 * time.Millisecond
	results2 <- 20 * time.Millisecond
	close(results2)

	printReport(&params, 1*time.Second, 2, 0, results2)
	output := params.String()
	assert.Contains(t, output, "Success: 2")
	assert.Contains(t, output, "Avg Latency:")
}

func TestRunBenchmark_Errors(t *testing.T) {
	// Closed server to force connection errors
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.Close() // Close immediately

	duration := 50 * time.Millisecond

	_, success, errors, _ := runBenchmark(ts.URL, 1, 0, duration, nil)

	assert.Equal(t, int64(0), success)
	assert.True(t, errors > 0)
}
