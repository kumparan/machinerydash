package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/kumparan/go-utils"
	"github.com/kumparan/machinerydash/dashboard"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

var (
	// machineryDash dashboardiface.Dashboard
	listStates = []string{tasks.StateFailure, tasks.StatePending, tasks.StateReceived, tasks.StateStarted, tasks.StateRetry, tasks.StateSuccess}
)

// Server :nodoc:
type Server struct {
	port          string
	viewsPath     string
	echo          *echo.Echo
	machineryDash dashboard.Dashboard
}

type cursorInfo struct {
	Cursor string
	Size   int64
}

type listTaskData struct {
	CurrentState   string
	EnableReEnqueu bool
	ListStates     []string
	TaskStates     []*dashboard.TaskWithSignature
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

	s.initRenderer()

	ec.GET("/", s.handleListAllTasksByState)
	ec.GET("/ping", s.handlePing)
	ec.POST("/reenqueue", s.handleReEnqueue)

	ec.Logger.Fatal(ec.Start(":" + s.port))
}

func (s *Server) initRenderer() error {
	bt, err := ioutil.ReadFile("views/index.html")
	if err != nil {
		return err
	}

	tpl := template.New("index.html")
	_, err = tpl.Parse(string(bt))
	if err != nil {
		return err
	}

	s.echo.Renderer = &htmlTemplate{templates: tpl}
	return nil
}

func (s *Server) handlePing(ec echo.Context) error {
	return ec.String(http.StatusOK, "pong")
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
		ListStates:     listStates,
		EnableReEnqueu: state == tasks.StateFailure,
		CurrentState:   state,
		TaskStates:     taskStates,
		cursorInfo: cursorInfo{
			Cursor: cursor,
			Size:   size,
		},
	}

	return ec.Render(http.StatusOK, "index.html", data)
}

func (s *Server) handleReEnqueue(ec echo.Context) error {
	sig := tasks.Signature{}
	req := struct {
		Signature string `json:"signature"`
	}{}
	err := ec.Bind(&req)
	if err != nil {
		logrus.Errorf("failed to parse request: %w", err)
		return ec.JSON(http.StatusBadRequest, fmtErr("invalid request"))
	}

	err = json.Unmarshal([]byte(req.Signature), &sig)
	if err != nil {
		logrus.Errorf("failed to unmarshal request: %w", err)
		return ec.JSON(http.StatusBadRequest, fmtErr("invalid request"))
	}

	err = s.machineryDash.ReEnqueueTask(&sig)
	if err != nil {
		logrus.Errorf("failed to ReEnqueueTask: %w", err)
		return ec.JSON(http.StatusInternalServerError, fmtErr("failed to reenqueue task"))
	}

	return ec.JSON(http.StatusOK, map[string]string{"message": "ok"})
}

func fmtErr(msg string) map[string]string {
	return map[string]string{"error": msg}
}
