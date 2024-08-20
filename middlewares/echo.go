package middlewares

import (
	"fmt"
	"strings"
	"time"

	"github.com/datadog/apm_tutorial_golang/tracer"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type stackTracer interface {
    StackTrace() errors.StackTrace
}

func extractFirstCaller(lines []string) string {
    for _, line := range lines {
        if strings.Contains(line, ".go") {
            return strings.TrimSpace(line)
        }
    }
    return "unknown"
}

func EchoLogger(logger *zap.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:           true,
		LogStatus:        true,
		LogError:         true,
		LogLatency:       true,
		LogRequestID:     true,
		LogMethod:        true,
		LogRemoteIP:      true,
		LogUserAgent:     true,
		LogContentLength: true,
		LogProtocol:      true,
		HandleError:      true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger := tracer.WithTrace(c.Request().Context(), logger)
            fields := []zap.Field{
                zap.String("request_id", v.RequestID),
                zap.String("method", v.Method),
                zap.String("uri", v.URI),
                zap.String("remote_ip", v.RemoteIP),
                zap.String("user_agent", v.UserAgent),
                zap.Duration("latency_ms", v.Latency*1000),
                zap.String("protocol", v.Protocol),
                zap.String("content_length", v.ContentLength),
                zap.String("start_time", v.StartTime.Format(time.RFC3339)),
                zap.Int("status", v.Status),
			}
			if v.Error == nil {
				logger.Info(fmt.Sprintf("%s success", v.URI), fields...)
			} else {
				var stacktrace string
				var caller string
                if err, ok := v.Error.(stackTracer); ok {
					stacktrace = strings.TrimSpace(fmt.Sprintf("%+v", err.StackTrace()))
					caller = extractFirstCaller(strings.Split(stacktrace, "\n"))
				    logger = logger.WithOptions(zap.WithCaller(false), zap.AddStacktrace(zapcore.InvalidLevel))
                }

                if caller != "" {
                    fields = append(fields, zap.String("caller", caller))
                }

                if stacktrace != "" {
                    fields = append(fields, zap.String("stacktrace", stacktrace))
                }

                if strings.Contains(v.Error.Error(), "not found") {
                    logger.Info(fmt.Sprintf("%s error: %v", v.URI, v.Error), fields...)
                } else {
                    fields = append(fields, zap.Error(v.Error))
                    logger.Error(fmt.Sprintf("%s error: %v", v.URI, v.Error), fields...)
                }
			}
			return nil
		},
	})
}
