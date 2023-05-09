package logging

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
)

func GraylogMiddlewareLogger() gin.HandlerFunc {
	return StructuredLogger()
}

// StructuredLogger logs a gin HTTP request in JSON format. Allows to set the
// logger for testing purposes.
func StructuredLogger() gin.HandlerFunc {

	return func(c *gin.Context) {
		start := time.Now() // Start timer
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Fill the params
		param := gin.LogFormatterParams{}

		param.TimeStamp = time.Now() // Stop timer
		param.Latency = param.TimeStamp.Sub(start)
		if param.Latency > time.Minute {
			param.Latency = param.Latency.Truncate(time.Second)
		}

		param.ClientIP = c.ClientIP()
		param.Method = c.Request.Method
		param.StatusCode = c.Writer.Status()
		param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
		param.BodySize = c.Writer.Size()
		if raw != "" {
			path = path + "?" + raw
		}
		param.Path = path

		// Log using the params
		fields := logrus.Fields{
			"_client_id":   param.ClientIP,
			"_method":      param.Method,
			"_status_code": param.StatusCode,
			"_body_size":   param.BodySize,
			"_path":        param.Path,
			"_latency":     param.Latency.String(),
		}
		logger := GetLoggerWithFields(fields)

		if c.Writer.Status() >= 500 {
			logger.Error(param.ErrorMessage)
		} else {
			logger.Info(param.Path)
		}
	}
}
