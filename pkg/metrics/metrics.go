package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/labstack/echo/v4"
)

type routeMetrics struct {
	totalRequests uint64
	totalLatency  uint64 // nanoseconds
}

var (
	totalRequests uint64
	inFlight      int64

	mu      sync.RWMutex
	metrics = make(map[string]map[int]*routeMetrics) // method -> status -> metrics
)

type metricRow struct {
	method string
	status int
	count  uint64
	sumNs  uint64
}

// Middleware records basic HTTP metrics for Prometheus scrapers.
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			atomic.AddInt64(&inFlight, 1)
			defer atomic.AddInt64(&inFlight, -1)

			start := time.Now()
			err := next(c)
			duration := time.Since(start)

			method := strings.ToUpper(c.Request().Method)
			status := c.Response().Status

			recordMetrics(method, status, duration)
			return err
		}
	}
}

// Handler returns an http.Handler that exposes metrics in Prometheus exposition format.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")

		// Global totals.
		fmt.Fprintf(w, "# HELP emplacc_http_requests_total Total number of HTTP requests processed.\n")
		fmt.Fprintf(w, "# TYPE emplacc_http_requests_total counter\n")
		fmt.Fprintf(w, "emplacc_http_requests_total %d\n", atomic.LoadUint64(&totalRequests))

		fmt.Fprintf(w, "# HELP emplacc_http_in_flight_requests Current number of in-flight HTTP requests.\n")
		fmt.Fprintf(w, "# TYPE emplacc_http_in_flight_requests gauge\n")
		fmt.Fprintf(w, "emplacc_http_in_flight_requests %d\n", atomic.LoadInt64(&inFlight))

		// Per-method/status metrics.
		rows := collectRows()

		fmt.Fprintf(w, "# HELP emplacc_http_requests_method_status_total Total number of HTTP requests partitioned by method and status.\n")
		fmt.Fprintf(w, "# TYPE emplacc_http_requests_method_status_total counter\n")
		for _, row := range rows {
			fmt.Fprintf(
				w,
				"emplacc_http_requests_method_status_total{method=%q,status=%q} %d\n",
				row.method,
				fmt.Sprint(row.status),
				row.count,
			)
		}

		fmt.Fprintf(w, "# HELP emplacc_http_request_duration_seconds_sum Total duration of HTTP requests in seconds partitioned by method and status.\n")
		fmt.Fprintf(w, "# TYPE emplacc_http_request_duration_seconds_sum counter\n")
		for _, row := range rows {
			durationSeconds := float64(row.sumNs) / float64(time.Second)
			fmt.Fprintf(
				w,
				"emplacc_http_request_duration_seconds_sum{method=%q,status=%q} %f\n",
				row.method,
				fmt.Sprint(row.status),
				durationSeconds,
			)
		}

		fmt.Fprintf(w, "# HELP emplacc_http_request_duration_seconds_count Number of latency samples recorded per method and status.\n")
		fmt.Fprintf(w, "# TYPE emplacc_http_request_duration_seconds_count counter\n")
		for _, row := range rows {
			fmt.Fprintf(
				w,
				"emplacc_http_request_duration_seconds_count{method=%q,status=%q} %d\n",
				row.method,
				fmt.Sprint(row.status),
				row.count,
			)
		}
	})
}

func recordMetrics(method string, status int, duration time.Duration) {
	atomic.AddUint64(&totalRequests, 1)

	mu.Lock()
	defer mu.Unlock()

	statusMap, ok := metrics[method]
	if !ok {
		statusMap = make(map[int]*routeMetrics)
		metrics[method] = statusMap
	}

	entry, ok := statusMap[status]
	if !ok {
		entry = &routeMetrics{}
		statusMap[status] = entry
	}

	entry.totalRequests++
	entry.totalLatency += uint64(duration)
}

func collectRows() []metricRow {
	mu.RLock()
	defer mu.RUnlock()

	rows := make([]metricRow, 0, len(metrics))
	for method, statusMap := range metrics {
		for status, entry := range statusMap {
			rows = append(rows, metricRow{
				method: method,
				status: status,
				count:  entry.totalRequests,
				sumNs:  entry.totalLatency,
			})
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].method == rows[j].method {
			return rows[i].status < rows[j].status
		}
		return rows[i].method < rows[j].method
	})

	return rows
}
