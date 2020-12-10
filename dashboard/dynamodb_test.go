package dashboard

import (
	"errors"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/RichardKnop/machinery/v1/backends/result"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/stretchr/testify/assert"
)

var jsonSignature = `{
	"UUID": "3",
	"Name": "DLQTaskCreateComment",
	"RoutingKey": "dlq-comment-service",
	"ETA": "2020-12-10T07:53:14.436882456Z",
	"Args": [
		{
			"Name": "userID",
			"Type": "int64",
			"Value": 1607416299930351600
		}
	],
	"RetryTimeout": 8
}`

func Test_FindAllTasksByState(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		dynamodbClient := &dynamodbClientMock{}
		machineryServer := &machineryServerMock{}
		dyn := &DynamoDB{
			cnf: &config.Config{
				DynamoDB: &config.DynamoDBConfig{},
			},
			client: dynamodbClient,
			server: machineryServer,
		}

		monkey.PatchInstanceMethod(reflect.TypeOf(dynamodbClient), "Query", func(*dynamodbClientMock, *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
			return nil, nil
		})

		res, cursor, err := dyn.FindAllTasksByState("", "", true, 10)
		assert.NoError(t, err)
		assert.Nil(t, res)
		assert.Empty(t, cursor)
	})

	t.Run("ok", func(t *testing.T) {
		dynamodbClient := &dynamodbClientMock{}
		machineryServer := &machineryServerMock{}
		dyn := &DynamoDB{
			cnf: &config.Config{
				DynamoDB: &config.DynamoDBConfig{},
			},
			client: dynamodbClient,
			server: machineryServer,
		}

		lasEvaluatedKey := map[string]*dynamodb.AttributeValue{
			"State":    {S: aws.String("FAILURE")},
			"TaskUUID": {S: aws.String("3")},
		}

		queryResult := &dynamodb.QueryOutput{
			LastEvaluatedKey: lasEvaluatedKey,
			Items: []map[string]*dynamodb.AttributeValue{
				{
					"State":    {S: aws.String("FAILURE")},
					"TaskUUID": {S: aws.String("1")},
				},
				{
					"State":    {S: aws.String("FAILURE")},
					"TaskUUID": {S: aws.String("3")},
				},
			},
		}
		monkey.PatchInstanceMethod(reflect.TypeOf(dynamodbClient), "Query", func(*dynamodbClientMock, *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
			return queryResult, nil
		})

		expectedCursor, err := encodeB64LastEvaluatedKey(lasEvaluatedKey)
		assert.NoError(t, err)

		res, cursor, err := dyn.FindAllTasksByState(tasks.StateFailure, "", true, 1)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(res))
		assert.Equal(t, expectedCursor, cursor)
	})
}

func Test_Rerun(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		dynamodbClient := &dynamodbClientMock{}
		machineryServer := &machineryServerMock{}
		dyn := &DynamoDB{
			cnf: &config.Config{
				DynamoDB: &config.DynamoDBConfig{},
			},
			client: dynamodbClient,
			server: machineryServer,
		}

		pg := monkey.PatchInstanceMethod(reflect.TypeOf(dyn), "FindTaskByUUID", func(*DynamoDB, string) (*TaskWithSignature, error) {
			return &TaskWithSignature{TaskUUID: "3", State: "FAILURE", Signature: jsonSignature}, nil
		})
		defer pg.Unpatch()

		pg2 := monkey.PatchInstanceMethod(reflect.TypeOf(machineryServer), "SendTask", func(*machineryServerMock, *tasks.Signature) (*result.AsyncResult, error) {
			return nil, nil
		})
		defer pg2.Unpatch()

		err := dyn.RerunTask("3")
		assert.NoError(t, err)
	})

	t.Run("handle GetItem error", func(t *testing.T) {
		dynamodbClient := &dynamodbClientMock{}
		machineryServer := &machineryServerMock{}
		dyn := &DynamoDB{
			cnf: &config.Config{
				DynamoDB: &config.DynamoDBConfig{},
			},
			client: dynamodbClient,
			server: machineryServer,
		}

		pg := monkey.PatchInstanceMethod(reflect.TypeOf(dyn), "FindTaskByUUID", func(*DynamoDB, string) (*TaskWithSignature, error) {
			return nil, errors.New("gotcha")
		})
		defer pg.Unpatch()

		err := dyn.RerunTask("3")
		assert.Error(t, err)
	})

	t.Run("handle SendTask error", func(t *testing.T) {
		dynamodbClient := &dynamodbClientMock{}
		machineryServer := &machineryServerMock{}
		dyn := &DynamoDB{
			cnf: &config.Config{
				DynamoDB: &config.DynamoDBConfig{},
			},
			client: dynamodbClient,
			server: machineryServer,
		}

		sig := jsonSignature
		queryResult := &dynamodb.GetItemOutput{
			Item: map[string]*dynamodb.AttributeValue{
				"State":     {S: aws.String("FAILURE")},
				"Signature": {S: aws.String(sig)},
				"TaskUUID":  {S: aws.String("3")},
			},
		}
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(dynamodbClient), "GetItem", func(*dynamodbClientMock, *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
			return queryResult, nil
		})
		defer pg.Unpatch()

		pg2 := monkey.PatchInstanceMethod(reflect.TypeOf(machineryServer), "SendTask", func(*machineryServerMock, *tasks.Signature) (*result.AsyncResult, error) {
			return nil, errors.New("gotcha")
		})
		defer pg2.Unpatch()

		err := dyn.RerunTask("3")
		assert.Error(t, err)
	})
}

func Test_FindTaskByUUID(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		dynamodbClient := &dynamodbClientMock{}
		machineryServer := &machineryServerMock{}
		dyn := &DynamoDB{
			cnf: &config.Config{
				DynamoDB: &config.DynamoDBConfig{},
			},
			client: dynamodbClient,
			server: machineryServer,
		}

		sig := jsonSignature
		queryResult := &dynamodb.GetItemOutput{
			Item: map[string]*dynamodb.AttributeValue{
				"State":     {S: aws.String("FAILURE")},
				"Signature": {S: aws.String(sig)},
				"TaskUUID":  {S: aws.String("3")},
			},
		}
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(dynamodbClient), "GetItem", func(*dynamodbClientMock, *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
			return queryResult, nil
		})
		defer pg.Unpatch()

		res, err := dyn.FindTaskByUUID("3")
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("handle error", func(t *testing.T) {
		dynamodbClient := &dynamodbClientMock{}
		machineryServer := &machineryServerMock{}
		dyn := &DynamoDB{
			cnf: &config.Config{
				DynamoDB: &config.DynamoDBConfig{},
			},
			client: dynamodbClient,
			server: machineryServer,
		}

		pg := monkey.PatchInstanceMethod(reflect.TypeOf(dynamodbClient), "GetItem", func(*dynamodbClientMock, *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
			return nil, errors.New("faild GetItem")
		})
		defer pg.Unpatch()

		res, err := dyn.FindTaskByUUID("3")
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}
