package db

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	"github.com/kumparan/machinerydash/config"
)

// NewDynamoDBClient create new dynamodb client to local instance or AWS instance
func NewDynamoDBClient() *dynamodb.DynamoDB {
	var sess *session.Session
	cfg := &aws.Config{
		Region:   aws.String(config.DynamoDBAWSRegion()),
		Endpoint: aws.String(config.DynamoDBHost()), // set this value when using local dynamodb
		Credentials: credentials.NewStaticCredentials(config.DynamoDBAWSAccessKey(),
			config.DynamoDBAWSSecretAccess(), ""),
	}
	sess = session.Must(session.NewSession(cfg))
	return dynamodb.New(sess)
}

// EnableDynamoDBTTL enable dynamodb ttl for a given table's attribute
func EnableDynamoDBTTL(client *dynamodb.DynamoDB, tableName string, attributeName string) error {
	desc, err := client.DescribeTimeToLive(&dynamodb.DescribeTimeToLiveInput{
		TableName: &tableName,
	})
	if err != nil {
		log.WithField("tableName", tableName).Fatal(err)
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

// AddStateIndex ..
func AddStateIndex(client *dynamodb.DynamoDB, tableName, attributeName string) {
	out, err := client.UpdateTable(&dynamodb.UpdateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("State"),
				AttributeType: aws.String("S"),
			},
		},
		TableName: aws.String(tableName),
		GlobalSecondaryIndexUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{
			{
				Create: &dynamodb.CreateGlobalSecondaryIndexAction{
					Projection: &dynamodb.Projection{
						ProjectionType: aws.String("ALL"),
					},
					IndexName: aws.String("StateIndex"),
					KeySchema: []*dynamodb.KeySchemaElement{
						{
							AttributeName: aws.String("State"),
							KeyType:       aws.String("HASH"),
						},
					},
					ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(5),
						WriteCapacityUnits: aws.Int64(5),
					},
				},
			},
		},
	})
	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Println(out.String())
}
