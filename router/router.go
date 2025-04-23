package router

import (
	"context"
	"log/slog"

	"github.com/flyx-ai/heimdall/handlers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewRouter(ctx context.Context) *echo.Echo {
	e := echo.New()

	e.Use(setupLogging(ctx))

	return setupRoutes(e)
}

func setupRoutes(e *echo.Echo) *echo.Echo {
	apiV1 := e.Group("/api/v1")

	apiV1.GET("/up", handlers.HandleUp)
	apiV1.GET("/complete", handlers.HandleComplete)

	return e
}

func setupLogging(ctx context.Context) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			level := slog.LevelInfo
			attrs := []slog.Attr{
				slog.Int("status", v.Status),
				slog.String("uri", v.URI),
				slog.String("method", v.Method),
				slog.String("host", v.Host),
			}
			if v.Error != nil {
				attrs = append(attrs, slog.String("error", v.Error.Error()))
				level = slog.LevelError
			}

			slog.Default().LogAttrs(ctx, level, "incoming_request",
				attrs...,
			)

			return nil
		},
	})
}
