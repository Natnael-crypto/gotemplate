package middleware

import (
	"bytes"
	"gotemplate/pkg/logger"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// structuredLogger logs HTTP requests with Zap
func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now() // Start time of the request

		// Read the request body to log it, then put it back for the next handlers
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
			// Restore the io.ReadCloser to its original state
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Process the request
		c.Next()

		// Log request details after processing
		duration := time.Since(start)      // Duration of the request
		status := c.Writer.Status()        // HTTP status code of the response
		path := c.Request.URL.Path         // Request URL path
		method := c.Request.Method         // HTTP method
		clientIP := c.ClientIP()           // Client IP address
		userAgent := c.Request.UserAgent() // User-Agent header
		responseSize := c.Writer.Size()    // Response body size

		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("ip", clientIP),
			zap.String("user_agent", userAgent),
			zap.Int("response_size", responseSize),
		}

		// Log request body if present
		if len(bodyBytes) > 0 {
			fields = append(fields, zap.ByteString("request_body", bodyBytes))
		}

		// Log error if any was set by other handlers
		if len(c.Errors) > 0 {
			// Append all errors
			for _, e := range c.Errors {
				fields = append(fields, zap.Error(e))
			}
			logger.Error("HTTP Request Error", fields...)
		} else {
			// Log based on status code
			if status >= 500 {
				logger.Error("HTTP Request Server Error", fields...)
			} else if status >= 400 {
				logger.Warn("HTTP Request Client Error", fields...)
			} else {
				logger.Info("HTTP Request", fields...)
			}
		}
	}
}
