package notes

import (
	"context"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type Logic interface {
	GetAllNotes(ctx context.Context) ([]Note, error)
	GetNote(ctx context.Context, id string) (Note, error)
	CreateNote(ctx context.Context, description string, addDate bool) (Note, error)
	UpdateNote(ctx context.Context, id string, newDescription string) (Note, error)
	DeleteNote(ctx context.Context, id string) error
}

type Router struct {
	Logger *zap.Logger
	Logic  Logic
}

func (nr *Router) Register(e *echo.Echo) {
    e.GET("/notes", WrapWithSpanMiddleware(nr.GetAllNotes))               // GET /notes
    e.POST("/notes", WrapWithSpanMiddleware(nr.CreateNote))               // POST /notes
    e.GET("/notes/:noteID", WrapWithSpanMiddleware(nr.GetNoteByID))       // GET /notes/123
    e.PUT("/notes/:noteID", WrapWithSpanMiddleware(nr.UpdateNoteByID))    // PUT /notes/123
    e.DELETE("/notes/:noteID", WrapWithSpanMiddleware(nr.DeleteNoteByID)) // DELETE /notes/123

	e.POST("/notes/quit", func(c echo.Context) error {
		time.AfterFunc(1*time.Second, func() { os.Exit(0) })
		return c.String(http.StatusOK, "Goodbye\n")
	}) // Quits program
}

func (nr *Router) GetAllNotes(c echo.Context) error {
	ctx := c.Request().Context()
	notes, err := nr.Logic.GetAllNotes(ctx)
	if err != nil {
		return err
	}

	doLongRunningProcess(ctx)
	anotherProcess(ctx)

	return c.JSON(http.StatusOK, notes)
}

func (nr *Router) GetNoteByID(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("noteID")
	if strings.TrimSpace(id) == "" {
		return errors.New("noteID not specified")
	}
	note, err := nr.Logic.GetNote(ctx, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, note)
}

func (nr *Router) CreateNote(c echo.Context) error {
	ctx := c.Request().Context()
	desc := c.QueryParam("desc")
	addDate := strings.EqualFold(c.QueryParam("add_date"), "y")

	note, err := nr.Logic.CreateNote(ctx, desc, addDate)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, note)
}

func (nr *Router) UpdateNoteByID(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("noteID")
	desc := c.QueryParam("desc")
	note, err := nr.Logic.UpdateNote(ctx, id, desc)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, note)
}

func (nr *Router) DeleteNoteByID(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("noteID")
	err := nr.Logic.DeleteNote(ctx, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, "Deleted")
}

func WrapWithSpanMiddleware(h echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        handlerName := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
        handlerName = strings.TrimSuffix(handlerName, "-fm")
        span, ctx := tracer.StartSpanFromContext(c.Request().Context(), handlerName)
        defer span.Finish()
        c.SetRequest(c.Request().WithContext(ctx))
        return h(c)
    }
}