package main

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	dashboard "github.com/RichardKnop/machinery/v1/dashboard/dynamodb"
	dashboardiface "github.com/RichardKnop/machinery/v1/dashboard/iface"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

var machineryDash dashboardiface.Dashboard

func init() {
	var err error
	sess := session.Must(session.NewSession(&aws.Config{
		Region:   aws.String("asia"),
		Endpoint: aws.String("http://localhost:8000"),
	}))
	machineryDash, err = machinery.NewDashboard(&config.Config{
		Broker:        "redis://localhost:6379/3",
		ResultBackend: "http://localhost:8000",
		DynamoDB: &config.DynamoDBConfig{
			TaskStatesTable: "task_states",
			GroupMetasTable: "group_metas",
			Client:          dynamodb.New(sess),
		},
		DefaultQueue:    "commerce-service-dlq-worker",
		ResultsExpireIn: 3600 * 24 * 30, // 30 days
	})
	if err != nil {
		logrus.Fatal(err)
	}
}

type htmlTemplate struct {
	templates *template.Template
}

func (t *htmlTemplate) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()
	e.Renderer = &htmlTemplate{template.Must(template.ParseGlob("views/*.html"))}

	e.Static("/static", "public")

	e.GET("/", index)
	e.GET("/ping", ping)
	e.POST("/reenqueue", reEnqueue)

	e.Logger.Fatal(e.Start(":9000"))
}

func ping(ec echo.Context) error {
	return ec.String(http.StatusOK, "pong")
}

func index(ec echo.Context) error {
	taskStates, err := machineryDash.FindAllTasksByState(tasks.StateFailure)
	if err != nil {
		logrus.Error(err)
		return ec.JSON(http.StatusInternalServerError, map[string]string{
			"error": "something wrong",
		})
	}

	data := struct {
		TaskStates []*dashboard.TaskWithSignature
	}{taskStates}
	return ec.Render(http.StatusOK, "index.html", data)
}

func reEnqueue(ec echo.Context) error {
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

	err = machineryDash.ReEnqueueTask(&sig)
	if err != nil {
		logrus.Errorf("failed to ReEnqueueTask: %w", err)
		return ec.JSON(http.StatusInternalServerError, fmtErr("failed to reenqueue task"))
	}

	return ec.JSON(http.StatusOK, map[string]string{"message": "ok"})
}

func fmtErr(msg string) map[string]string {
	return map[string]string{"error": msg}
}
