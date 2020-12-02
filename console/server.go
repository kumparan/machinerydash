package console

import (
	"github.com/RichardKnop/machinery/v1"
	machineryConfig "github.com/RichardKnop/machinery/v1/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/kumparan/machinerydash/config"
	"github.com/kumparan/machinerydash/dashboard"
	"github.com/kumparan/machinerydash/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serverCMD = &cobra.Command{
	Use:   "server",
	Short: "run server",
	Long:  `This subcommand start the server`,
	Run:   runServer,
}

func init() {
	RootCmd.AddCommand(serverCMD)
}

func runServer(cmd *cobra.Command, args []string) {
	cfg := createMachineryCfg()

	machineryServer, err := machinery.NewServer(cfg)
	if err != nil {
		logrus.Fatal(err)
	}

	machineryDash := dashboard.NewDynamodb(cfg, machineryServer)
	srv := server.New(config.Port(), machineryDash)
	srv.Start()
}

func createDynamoDBSession() *session.Session {
	if config.IsLocalDynamodb() {
		return session.Must(session.NewSession(&aws.Config{
			Region:   aws.String(config.DynamodbRegion()),
			Endpoint: aws.String(config.DynamodbHost()),
		}))
	}

	return nil // TODO: handle for non local db
}

func createMachineryCfg() *machineryConfig.Config {
	return &machineryConfig.Config{
		Broker:        config.MachineryBrokerHost(),
		ResultBackend: config.DynamodbHost(),
		DynamoDB: &machineryConfig.DynamoDBConfig{
			TaskStatesTable: config.DynamodbTaskTable(),
			GroupMetasTable: config.DynamodbGroupTable(),
			Client:          dynamodb.New(createDynamoDBSession()),
		},
		DefaultQueue:    config.MachineryBrokerNamespace(), // use namespace as queueu
		ResultsExpireIn: config.MachineryResultExpiry(),
	}
}
