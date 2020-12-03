package console

import (
	"fmt"

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
	dynamoDBClient := dynamodb.New(createDynamoDBSession())
	cfg := &machineryConfig.Config{
		Broker:        config.MachineryBrokerHost(),
		ResultBackend: config.DynamodbHost(),
		DynamoDB: &machineryConfig.DynamoDBConfig{
			TaskStatesTable: config.DynamodbTaskTable(),
			GroupMetasTable: config.DynamodbGroupTable(),
			Client:          dynamoDBClient,
		},
		DefaultQueue:    config.MachineryBrokerNamespace(), // use namespace as queue
		ResultsExpireIn: config.MachineryResultExpiry(),
	}

	err := enableDynamoDBTTL(dynamoDBClient, cfg.DynamoDB.TaskStatesTable)
	if err != nil {
		logrus.Fatal(err)
	}

	err = enableDynamoDBTTL(dynamoDBClient, cfg.DynamoDB.GroupMetasTable)
	if err != nil {
		logrus.Fatal(err)
	}

	return cfg
}

func enableDynamoDBTTL(client *dynamodb.DynamoDB, tableName string) error {
	desc, err := client.DescribeTimeToLive(&dynamodb.DescribeTimeToLiveInput{
		TableName: &tableName,
	})
	if err != nil {
		logrus.Fatal(err)
	}

	isDynamoDBTTLEnabled := *desc.TimeToLiveDescription.TimeToLiveStatus == dynamodb.TimeToLiveStatusEnabled
	isDynamoDBTTLEnabling := *desc.TimeToLiveDescription.TimeToLiveStatus == dynamodb.TimeToLiveStatusEnabling
	if isDynamoDBTTLEnabled || isDynamoDBTTLEnabling {
		return nil
	}

	logrus.Infof("enabling ttl on table %s", tableName)
	out, err := client.UpdateTimeToLive(&dynamodb.UpdateTimeToLiveInput{
		TableName: &tableName,
		TimeToLiveSpecification: &dynamodb.TimeToLiveSpecification{
			Enabled:       aws.Bool(true),
			AttributeName: aws.String("TTL"),
		},
	})
	if err != nil {
		return fmt.Errorf("failed when updating ttl on table %s: %w", tableName, err)
	}

	if !*out.TimeToLiveSpecification.Enabled {
		return fmt.Errorf("failed to enable TTL on dynamodb table %s", tableName)
	}

	return nil
}
