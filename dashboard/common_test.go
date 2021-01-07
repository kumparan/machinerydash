package dashboard

import (
	"github.com/RichardKnop/machinery/v1/backends/result"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type dynamodbClientMock struct{}

func (d *dynamodbClientMock) Query(*dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	return nil, nil
}

func (d *dynamodbClientMock) GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	return nil, nil
}

type machineryServerMock struct{}

func (m *machineryServerMock) SendTask(signature *tasks.Signature) (*result.AsyncResult, error) {
	return nil, nil
}
