package server

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/kumparan/go-utils"
	"github.com/kumparan/machinerydash/dashboard"
	"github.com/labstack/echo/v4"
	"github.com/markbates/pkger"
	"github.com/sirupsen/logrus"
)

var (
	stateList = []string{tasks.StateFailure, tasks.StatePending, tasks.StateReceived, tasks.StateStarted, tasks.StateRetry, tasks.StateSuccess}
)

// Server :nodoc:
type Server struct {
	// viewsPath     string
	port          string
	echo          *echo.Echo
	machineryDash dashboard.Dashboard
}

type cursorInfo struct {
	Cursor string
	Size   int64
}

type listTaskData struct {
	CurrentState string
	EnableRerun  bool
	ListStates   []string
	TaskStates   []*dashboard.TaskWithSignature
	cursorInfo
}

// New :nodoc:
func New(port string, md dashboard.Dashboard) *Server {
	return &Server{
		port:          port,
		echo:          echo.New(),
		machineryDash: md,
	}
}

// Start :nodoc:
func (s *Server) Start() {
	ec := s.echo

	if err := s.initRenderer(); err != nil {
		logrus.Fatal(err)
	}

	ec.GET("/", s.handleListAllTasksByState)
	ec.GET("/ping", s.handlePing)
	ec.GET("/static/*", s.handleStatic)
	ec.POST("/rerun", s.handleRerun)

	ec.Logger.Fatal(ec.Start(":" + s.port))
}

func (s *Server) initRenderer() error {
	serverTemplate := template.New("")
	err := pkger.Walk("/views", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := pkger.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open %s from pkger: %w", path, err)
		}

		bt, err := ioutil.ReadAll(f)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		_, err = serverTemplate.New(info.Name()).Parse(string(bt))
		if err != nil {
			return fmt.Errorf("unable to parse template from %s: %w", path, err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.echo.Renderer = &htmlTemplate{templates: serverTemplate}
	return nil
}

func (s *Server) handlePing(ec echo.Context) error {
	return ec.String(http.StatusOK, "pong")
}

func (s *Server) handleStatic(c echo.Context) error {
	f, err := pkger.Open(path.Join("/public", c.Param("*")))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logrus.Error(err)
		}
		return c.NoContent(http.StatusNotFound)
	}

	bt, err := ioutil.ReadAll(f)
	if err != nil {
		logrus.Error(err)
		return c.String(http.StatusInternalServerError, "something wrong")
	}

	c.Response().Header().Set(echo.HeaderContentType, "text/css")
	return c.String(http.StatusOK, string(bt))
}

func (s *Server) handleListAllTasksByState(ec echo.Context) error {
	next := ec.QueryParam("next")
	prev := ec.QueryParam("prev")
	size := utils.StringToInt64(ec.QueryParam("size"))
	state := strings.ToUpper(ec.QueryParam("state"))

	if strings.TrimSpace(state) == "" {
		state = tasks.StateFailure
	}

	cursor := next
	if prev != "" {
		cursor = prev
	}

	taskStates, cursor, err := s.machineryDash.FindAllTasksByState(state, cursor, true, size)
	if err != nil {
		logrus.Error(err)
		return ec.JSON(http.StatusInternalServerError, map[string]string{
			"error": "something wrong",
		})
	}

	data := listTaskData{
		ListStates:   stateList,
		EnableRerun:  state == tasks.StateFailure,
		CurrentState: state,
		TaskStates:   taskStates,
		cursorInfo: cursorInfo{
			Cursor: cursor,
			Size:   size,
		},
	}

	return ec.Render(http.StatusOK, "index.html", data)
}

func (s *Server) handleRerun(ec echo.Context) error {
	req := struct {
		UUID string `json:"uuid"`
	}{}
	err := errors.Unwrap(ec.Bind(&req))
	if err != nil {
		logrus.Error(err)
		return ec.JSON(http.StatusBadRequest, fmtErr("invalid request"))
	}

	err = s.machineryDash.RerunTask(req.UUID)
	if err != nil {
		logrus.WithField("uuid", req.UUID).Error(err)
		return ec.JSON(http.StatusInternalServerError, fmtErr("failed to rerun task"))
	}

	return ec.JSON(http.StatusOK, map[string]string{"message": "ok"})
}

func fmtErr(msg string) map[string]string {
	return map[string]string{"error": msg}
}
