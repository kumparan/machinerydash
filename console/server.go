package console

import (
	"github.com/RichardKnop/machinery/v1"
	machineryConfig "github.com/RichardKnop/machinery/v1/config"
	"github.com/kumparan/machinerydash/config"
	"github.com/kumparan/machinerydash/dashboard"
	"github.com/kumparan/machinerydash/db"
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

func createMachineryCfg() *machineryConfig.Config {
	dynamoDBClient := db.NewDynamoDBClient()
	cfg := &machineryConfig.Config{
		Broker:        config.MachineryBrokerHost(),
		ResultBackend: config.DynamodbHost(),
		DynamoDB: &machineryConfig.DynamoDBConfig{
			TaskStatesTable: config.DynamoDBTaskTable(),
			GroupMetasTable: config.DynamoDBGroupTable(),
			Client:          dynamoDBClient,
		},
		DefaultQueue:    config.MachineryBrokerNamespace(), // use namespace as queue
		ResultsExpireIn: config.MachineryResultExpiry(),
	}

	err := db.EnableDynamoDBTTL(dynamoDBClient, cfg.DynamoDB.TaskStatesTable, "TTL")
	if err != nil {
		logrus.Fatal(err)
	}

	err = db.EnableDynamoDBTTL(dynamoDBClient, cfg.DynamoDB.GroupMetasTable, "TTL")
	if err != nil {
		logrus.Fatal(err)
	}

	return cfg
}
