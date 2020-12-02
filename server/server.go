package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/kumparan/machinerydash/dashboard"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

var (
	// machineryDash dashboardiface.Dashboard
	listState = []string{tasks.StateReceived, tasks.StatePending, tasks.StateStarted, tasks.StateRetry, tasks.StateSuccess, tasks.StateFailure}
)

// Server :nodoc:
type Server struct {
	port          string
	viewsPath     string
	echo          *echo.Echo
	machineryDash dashboard.Dashboard
}

type listTaskData struct {
	CurrentState   string
	EnableReEnqueu bool
	ListStates     []string
	TaskStates     []*dashboard.TaskWithSignature
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

	ec.Static("/static", "public")

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
	state := ec.QueryParam("state")
	if strings.TrimSpace(state) == "" {
		state = tasks.StateStarted
	}

	state = strings.ToUpper(state)
	taskStates, _, err := s.machineryDash.FindAllTasksByState(state, "", false, 10)
	if err != nil {
		logrus.Error(err)
		return ec.JSON(http.StatusInternalServerError, map[string]string{
			"error": "something wrong",
		})
	}

	data := listTaskData{
		ListStates:     listState,
		EnableReEnqueu: state == tasks.StateFailure,
		CurrentState:   state,
		TaskStates:     taskStates,
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
