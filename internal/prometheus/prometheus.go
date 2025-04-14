package prometheus

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func PrometheusMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())
		path := c.Path()

		httpRequestsTotal.WithLabelValues(
			c.Method(),
			path,
			status,
		).Inc()

		httpResponseTime.WithLabelValues(
			c.Method(),
			path,
		).Observe(duration)

		return err
	}
}
