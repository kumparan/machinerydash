package main

import (
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

	e.GET("/", index)
	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})
	e.Static("/static", "public")
	e.Logger.Fatal(e.Start(":9000"))
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
