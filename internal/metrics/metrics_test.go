package metrics

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestMiddlewareAndHandlerExposeLowCardinalityMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := New()
	r := gin.New()
	r.Use(m.Middleware())
	r.GET("/logs/:id", func(c *gin.Context) { c.Status(201) })

	req := httptest.NewRequest("GET", "/logs/42", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	require.Equal(t, 201, res.Code)
	m.RecordKafka("success")
	metricsRes := httptest.NewRecorder()
	m.Handler().ServeHTTP(metricsRes, httptest.NewRequest("GET", "/metrics", nil))

	body := metricsRes.Body.String()
	require.Contains(t, body, `logpulse_http_requests_total{method="GET",route="/logs/:id",status="201"} 1`)
	require.Contains(t, body, `logpulse_http_request_duration_seconds_count{method="GET",route="/logs/:id"} 1`)
	require.Contains(t, body, `logpulse_kafka_messages_total{result="success"} 1`)
	require.NotContains(t, strings.ToLower(body), "request_id")
}
