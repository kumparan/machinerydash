package dashboard

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/stretchr/testify/assert"
)

func Test_FindAllTasksByState(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		dynamodbClient := &dynamodbClientMock{}
		machineryServer := &machineryServerMock{}
		dyn := &Dynamodb{
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
		dyn := &Dynamodb{
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
