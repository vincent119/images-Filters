package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	targetURL   string
	concurrency int
	requests    int
	duration    time.Duration
)

func init() {
	flag.StringVar(&targetURL, "url", "http://localhost:8080/healthz", "Target URL")
	flag.IntVar(&concurrency, "c", 10, "Concurrency level")
	flag.IntVar(&requests, "n", 0, "Number of requests to run (0 for duration-based)")
	flag.DurationVar(&duration, "d", 10*time.Second, "Duration to run (if n is 0)")
}

func main() {
	flag.Parse()

	fmt.Printf("Benchmarking %s\n", targetURL)
	fmt.Printf("Concurrency: %d\n", concurrency)
	if requests > 0 {
		fmt.Printf("Requests: %d\n", requests)
	} else {
		fmt.Printf("Duration: %v\n", duration)
	}

	totalDuration, successCount, errorCount, results := runBenchmark(targetURL, concurrency, requests, duration, nil)
	printReport(os.Stdout, totalDuration, successCount, errorCount, results)
}

func runBenchmark(target string, conc int, numReq int, dur time.Duration, serverWait chan struct{}) (time.Duration, int64, int64, chan time.Duration) {
	start := time.Now()
	var wg sync.WaitGroup
	// Buffer enough results
	results := make(chan time.Duration, 100000)
	var successCount int64
	var errorCount int64

	// Adjust total requests or use timer
	stopChan := make(chan struct{})
	if numReq == 0 {
		time.AfterFunc(dur, func() {
			close(stopChan)
		})
	}

	work := func() {
		defer wg.Done()
		for {
			select {
			case <-stopChan:
				return
			default:
				if serverWait != nil {
					// wait for server signal if testing hook needed,
					// but for now simple http.Get is fine.
				}
			}

			reqStart := time.Now()
			resp, err := http.Get(target)
			latency := time.Since(reqStart)

			if err != nil {
				atomic.AddInt64(&errorCount, 1)
			} else {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					atomic.AddInt64(&successCount, 1)
					select {
					case results <- latency:
					default:
						// Buffer full
					}
				} else {
					atomic.AddInt64(&errorCount, 1)
				}
			}

			if numReq > 0 {
				// Simple implementation limitation mentioned in original code
				return
			}
		}
	}

	wg.Add(conc)
	for i := 0; i < conc; i++ {
		go work()
	}

	wg.Wait()
	total := time.Since(start)
	close(results)

	return total, successCount, errorCount, results
}

func printReport(w io.Writer, totalDuration time.Duration, success, errors int64, results chan time.Duration) {
	fmt.Fprintf(w, "\n--- Report ---\n")
	fmt.Fprintf(w, "Total Duration: %v\n", totalDuration)
	fmt.Fprintf(w, "Total Requests: %d\n", success+errors)
	fmt.Fprintf(w, "Success: %d\n", success)
	fmt.Fprintf(w, "Errors: %d\n", errors)
	rps := 0.0
	if totalDuration.Seconds() > 0 {
		rps = float64(success+errors) / totalDuration.Seconds()
	}
	fmt.Fprintf(w, "RPS: %.2f\n", rps)

	var latencies []time.Duration
	for l := range results {
		latencies = append(latencies, l)
	}

	if len(latencies) > 0 {
		var totalLat time.Duration
		var maxLat time.Duration
		minLat := latencies[0]

		for _, l := range latencies {
			totalLat += l
			if l > maxLat {
				maxLat = l
			}
			if l < minLat {
				minLat = l
			}
		}
		avgLat := totalLat / time.Duration(len(latencies))
		fmt.Fprintf(w, "Avg Latency: %v\n", avgLat)
		fmt.Fprintf(w, "Min Latency: %v\n", minLat)
		fmt.Fprintf(w, "Max Latency: %v\n", maxLat)
	}
}
