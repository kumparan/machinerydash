package dashboard

import (
	"github.com/RichardKnop/machinery/v1/backends/result"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// Dashboard :noodc:
type Dashboard interface {
	FindAllTasksByState(state, cursor string, asc bool, size int64) (taskStates []*TaskWithSignature, next string, err error)
	RerunTask(sig *tasks.Signature) error
}

type machineryServer interface {
	SendTask(signature *tasks.Signature) (*result.AsyncResult, error)
}

type dynamoDBClient interface {
	Query(*dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
}
