package calendar

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func GetDate(c echo.Context) error {
	span, _ := tracer.StartSpanFromContext(c.Request().Context(), "GetDate")
	defer span.Finish()

	val := rand.Intn(365)
	date := time.Now().AddDate(0, 0, val)
	ranDate, marshError := json.Marshal(date.Format("2006-01-02"))
	if marshError != nil {
		log.Fatalln(marshError)
	}
	log.Println(string(ranDate))

	return c.JSONBlob(http.StatusOK, ranDate)
}
