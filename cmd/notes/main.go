package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"fmt"

	"go.uber.org/zap"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
	echotrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/labstack/echo.v4"
	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/datadog/apm_tutorial_golang/notes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"

	dd_logrus "gopkg.in/DataDog/dd-trace-go.v1/contrib/sirupsen/logrus"
)

func main() {
	tracer.Start(tracer.WithRuntimeMetrics())
	defer tracer.Stop()

    // Optional: Change log format to use JSON (Cf. Go Log Collection)
    logrus.SetFormatter(&logrus.JSONFormatter{})

    // Add Datadog context log hook
    logrus.AddHook(&dd_logrus.DDContextLogHook{}) 

	logger, _ := zap.NewDevelopment()
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
	e.Use(middleware.Logger())
	e.Use(echotrace.Middleware(echotrace.WithServiceName("notes")))

	nr.Register(e) // Adjusted to work with Echo

	log.Fatal(e.Start(":8080"))
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
