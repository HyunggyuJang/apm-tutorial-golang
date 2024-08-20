package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"runtime"

	log "github.com/datadog/apm_tutorial_golang/logger"
	tracing "github.com/datadog/apm_tutorial_golang/tracer"
	"go.uber.org/zap"

	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
	echotrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/labstack/echo.v4"
	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/datadog/apm_tutorial_golang/middlewares"
	"github.com/datadog/apm_tutorial_golang/notes"
	"github.com/labstack/echo/v4"
	"github.com/mattn/go-sqlite3"
)

func main() {
	tracer.Start(tracer.WithRuntimeMetrics())
	defer tracer.Stop()

	logger := log.New()
	ctx := context.Background()
	_, ctx = tracer.StartSpanFromContext(ctx, getCurrentFunctionName())
	logger = tracing.WithTrace(ctx, logger)
	logger.Debug("Starting notes service")
	logger.Debug("Starting from port 8080")

	db := setupDB(logger)
	defer db.Close()

	client := http.DefaultClient
	// Creates span with resource name equal to http Method and path
	client = httptrace.WrapClient(client, httptrace.RTWithResourceNamer(func(req *http.Request) string {
		return fmt.Sprintf("%s %s", req.Method, req.URL.Path)
	}))

	host, found := os.LookupEnv("CALENDAR_HOST")
	if !found || host == "" {
		host = "localhost"
	}

	logic := &notes.LogicImpl{
		DB:           db,
		Logger:       logger,
		Client:       client,
		CalendarHost: host,
	}

	nr := notes.Router{
		Logger: logger,
		Logic:  logic,
	}

	e := echo.New()
	e.Use(echotrace.Middleware(echotrace.WithServiceName("notes")))
	e.Use(middlewares.EchoLogger(logger))

	nr.Register(e) // Adjusted to work with Echo

	logger.Fatal("Error starting server", zap.Error(e.Start(":8080")))
}

func setupDB(logger *zap.Logger) *sql.DB {
	sqltrace.Register("sqlite3", &sqlite3.SQLiteDriver{}, sqltrace.WithServiceName("db"))
	db, err := sqltrace.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		logger.Fatal("error setting up database", zap.Error(err))
	}

	sts := ` DROP TABLE IF EXISTS notes;
			CREATE TABLE notes(id INTEGER PRIMARY KEY, description TEXT);`
	_, err = db.Exec(sts)
	if err != nil {
		logger.Fatal("error creating schema", zap.Error(err))
	}
	return db
}

func getCurrentFunctionName() string {
    pc, _, _, ok := runtime.Caller(1)
    if !ok {
        return "unknown"
    }
    fn := runtime.FuncForPC(pc)
    if fn == nil {
        return "unknown"
    }
    return fn.Name()
}