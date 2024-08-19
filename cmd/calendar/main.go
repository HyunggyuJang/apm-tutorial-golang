package main

import (
	"log"
	"os"
	"time"

	"github.com/datadog/apm_tutorial_golang/calendar"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echotrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/labstack/echo.v4"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func main() {
	tracer.Start(tracer.WithRuntimeMetrics())
	defer tracer.Stop()

	log.Printf("Starting from port 9090")

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(echotrace.Middleware(echotrace.WithServiceName("calendar")))

	e.GET("/calendar", calendar.GetDate)
	e.POST("/calendar/quit", func(c echo.Context) error {
		time.AfterFunc(1*time.Second, func() { os.Exit(0) })
		return c.String(200, "Goodbye\n")
	})

	log.Fatal(e.Start(":9090"))
}
