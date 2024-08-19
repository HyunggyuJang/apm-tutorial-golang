package notes

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/labstack/echo/v4"
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
	e.GET("/notes", nr.makeSpanMiddleware("GetAllNotes", nr.GetAllNotes))               // GET /notes
	e.POST("/notes", nr.makeSpanMiddleware("CreateNote", nr.CreateNote))                // POST /notes
	e.GET("/notes/:noteID", nr.makeSpanMiddleware("GetNote", nr.GetNoteByID))           // GET /notes/123
	e.PUT("/notes/:noteID", nr.makeSpanMiddleware("UpdateNote", nr.UpdateNoteByID))     // PUT /notes/123
	e.DELETE("/notes/:noteID", nr.makeSpanMiddleware("DeleteNote", nr.DeleteNoteByID))  // DELETE /notes/123

	e.POST("/notes/quit", func(c echo.Context) error {
		time.AfterFunc(1*time.Second, func() { os.Exit(0) })
		return c.String(http.StatusOK, "Goodbye\n")
	}) // Quits program
}

func (nr *Router) makeSpanMiddleware(name string, h echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := tracer.StartSpanFromContext(c.Request().Context(), name)
		defer span.Finish()
		c.SetRequest(c.Request().WithContext(ctx))
		return h(c)
	}
}

func reportError(err error, category string, c echo.Context) error {
	msg := struct {
		Category string `json:"category"`
		Message  string `json:"message"`
	}{
		Category: category,
		Message:  err.Error(),
	}

	return c.JSON(http.StatusInternalServerError, msg)
}

func reportInputError(message string, c echo.Context) error {
	msg := struct {
		Category string `json:"category"`
		Message  string `json:"message"`
	}{
		Category: message,
		Message:  "invalid input",
	}

	return c.JSON(http.StatusBadRequest, msg)
}

func (nr *Router) GetAllNotes(c echo.Context) error {
	ctx := c.Request().Context()
	notes, err := nr.Logic.GetAllNotes(ctx)
	if err != nil {
		return reportError(err, "GetAllNotes", c)
	}

	doLongRunningProcess(ctx)
	anotherProcess(ctx)

	return c.JSON(http.StatusOK, notes)
}

func (nr *Router) GetNoteByID(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("noteID")
	if strings.TrimSpace(id) == "" {
		return reportInputError("noteID not specified", c)
	}
	note, err := nr.Logic.GetNote(ctx, id)
	if err != nil {
		return reportError(err, "GetNoteByID", c)
	}

	return c.JSON(http.StatusOK, note)
}

func (nr *Router) CreateNote(c echo.Context) error {
	ctx := c.Request().Context()
	desc := c.QueryParam("desc")
	addDate := strings.EqualFold(c.QueryParam("add_date"), "y")

	note, err := nr.Logic.CreateNote(ctx, desc, addDate)
	if err != nil {
		return reportError(err, "CreateNote", c)
	}

	return c.JSON(http.StatusCreated, note)
}

func (nr *Router) UpdateNoteByID(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("noteID")
	desc := c.QueryParam("desc")
	note, err := nr.Logic.UpdateNote(ctx, id, desc)
	if err != nil {
		return reportError(err, "UpdateNoteByID", c)
	}

	return c.JSON(http.StatusOK, note)
}

func (nr *Router) DeleteNoteByID(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("noteID")
	err := nr.Logic.DeleteNote(ctx, id)
	if err != nil {
		return reportError(err, "DeleteNoteByID", c)
	}

	return c.JSON(http.StatusOK, "Deleted")
}
