package logging

import (
	ginZap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"time"
)

// GraylogMiddlewareLogger logs a gin HTTP request in JSON format. Allows to set the
// logger for testing purposes.
//func GraylogMiddlewareLogger() gin.HandlerFunc {
//
//	// Logrus will send logs to Graylog
//	return func(c *gin.Context) {
//		start := time.Now() // Start timer
//		path := c.Request.URL.Path
//		raw := c.Request.URL.RawQuery
//
//		// Process request
//		c.Next()
//
//		// Fill the params
//		param := gin.LogFormatterParams{}
//
//		param.TimeStamp = time.Now() // Stop timer
//		param.Latency = param.TimeStamp.Sub(start)
//		if param.Latency > time.Minute {
//			param.Latency = param.Latency.Truncate(time.Second)
//		}
//
//		param.ClientIP = c.ClientIP()
//		param.Method = c.Request.Method
//		param.StatusCode = c.Writer.Status()
//		param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
//		param.BodySize = c.Writer.Size()
//		if raw != "" {
//			path = path + "?" + raw
//		}
//		param.Path = path
//
//		// Log using the params
//		fields := []zap.Field{
//			zap.String("_client_id", param.ClientIP),
//			zap.String("_method", param.Method),
//			zap.Int("_status_code", param.StatusCode),
//			zap.Int("_body_size", param.BodySize),
//			zap.String("_path", param.Path),
//			zap.Duration("_latency", param.Latency),
//		}
//		logger := GetLogger()
//
//		if c.Writer.Status() >= 500 {
//			logger.Error(param.ErrorMessage, fields...)
//		} else {
//			logger.Info(param.Path, fields...)
//		}
//	}
//}

func ZapMiddlewareLogger(router *gin.Engine) {
	l := GetLogger()
	router.Use(ginZap.GinzapWithConfig(l.Logger.Logger, &ginZap.Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
		SkipPaths:  nil,
		TraceID:    true,
		// extra
		Context: nil,
	}))
	router.Use(ginZap.RecoveryWithZap(l.Logger.Logger, true))
}
