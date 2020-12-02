package main

import (
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	dashboardiface "github.com/RichardKnop/machinery/v1/dashboard/iface"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/kumparan/machinerydash/server"
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

func main() {
	srv := server.New("9000", machineryDash)
	srv.Start()
}
