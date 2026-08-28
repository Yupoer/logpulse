package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrics keeps a small, in-memory set of Prometheus counters.
// Labels are route templates, not user-provided paths, so cardinality stays low.
type Metrics struct {
	mu        sync.Mutex
	requests  map[httpKey]uint64
	durations map[routeKey]duration
	kafka     map[string]uint64
}

type httpKey struct {
	method string
	route  string
	status string
}

type routeKey struct {
	method string
	route  string
}

type duration struct {
	count uint64
	sum   float64
}

func New() *Metrics {
	return &Metrics{
		requests:  make(map[httpKey]uint64),
		durations: make(map[routeKey]duration),
		kafka:     make(map[string]uint64),
	}
}

// Middleware records one request after all other middleware and handlers finish.
func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "/unknown"
		}
		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())

		m.mu.Lock()
		m.requests[httpKey{method: method, route: route, status: status}]++
		key := routeKey{method: method, route: route}
		value := m.durations[key]
		value.count++
		value.sum += time.Since(started).Seconds()
		m.durations[key] = value
		m.mu.Unlock()
	}
}

func (m *Metrics) RecordKafka(result string) {
	m.RecordKafkaBatch(1, result)
}

func (m *Metrics) RecordKafkaBatch(size int, result string) {
	if size <= 0 {
		return
	}
	m.mu.Lock()
	m.kafka[result] += uint64(size)
	m.mu.Unlock()
}

// Handler writes the small Prometheus text format without adding a dependency.
func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintln(w, "# HELP logpulse_http_requests_total Total HTTP requests handled.")
		fmt.Fprintln(w, "# TYPE logpulse_http_requests_total counter")
		for _, key := range sortedHTTPKeys(m.requests) {
			fmt.Fprintf(w, "logpulse_http_requests_total{method=\"%s\",route=\"%s\",status=\"%s\"} %d\n",
				escape(key.method), escape(key.route), escape(key.status), m.requests[key])
		}

		fmt.Fprintln(w, "# HELP logpulse_http_request_duration_seconds HTTP request duration in seconds.")
		fmt.Fprintln(w, "# TYPE logpulse_http_request_duration_seconds summary")
		for _, key := range sortedRouteKeys(m.durations) {
			value := m.durations[key]
			labels := fmt.Sprintf("method=\"%s\",route=\"%s\"", escape(key.method), escape(key.route))
			fmt.Fprintf(w, "logpulse_http_request_duration_seconds_count{%s} %d\n", labels, value.count)
			fmt.Fprintf(w, "logpulse_http_request_duration_seconds_sum{%s} %.6f\n", labels, value.sum)
		}

		fmt.Fprintln(w, "# HELP logpulse_kafka_messages_total Kafka messages by processing result.")
		fmt.Fprintln(w, "# TYPE logpulse_kafka_messages_total counter")
		results := make([]string, 0, len(m.kafka))
		for result := range m.kafka {
			results = append(results, result)
		}
		sort.Strings(results)
		for _, result := range results {
			fmt.Fprintf(w, "logpulse_kafka_messages_total{result=\"%s\"} %d\n", escape(result), m.kafka[result])
		}
	})
}

func sortedHTTPKeys(values map[httpKey]uint64) []httpKey {
	keys := make([]httpKey, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].method != keys[j].method {
			return keys[i].method < keys[j].method
		}
		if keys[i].route != keys[j].route {
			return keys[i].route < keys[j].route
		}
		return keys[i].status < keys[j].status
	})
	return keys
}

func sortedRouteKeys(values map[routeKey]duration) []routeKey {
	keys := make([]routeKey, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].method != keys[j].method {
			return keys[i].method < keys[j].method
		}
		return keys[i].route < keys[j].route
	})
	return keys
}

func escape(value string) string {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`).Replace(value)
}
