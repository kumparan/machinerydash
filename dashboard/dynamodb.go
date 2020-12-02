package dashboard

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/log"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// Dynamodb monitor tasks
type Dynamodb struct {
	cnf    *config.Config
	client dynamoDBClient
	server machineryServer
}

// TaskWithSignature :nodoc:
type TaskWithSignature struct {
	TaskUUID  string `bson:"task_uuid"`
	State     string `bson:"state"`
	TaskName  string `bson:"task_name"`
	Signature string `bson:"signature"`
	CreatedAt string `bson:"created_at"`
	Error     string `bson:"error"`
}

// NewDynamodb :nodoc:
func NewDynamodb(cnf *config.Config, srv machineryServer) Dashboard {
	dash := &Dynamodb{
		cnf:    cnf,
		server: srv,
	}

	if cnf.DynamoDB != nil && cnf.DynamoDB.Client != nil {
		dash.client = cnf.DynamoDB.Client
	} else {
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))
		dash.client = dynamodb.New(sess)
	}

	return dash
}

// FindAllTasksByState :nodoc:
// cursor e.g. "prev" & "next" are base64 encoded LastEvaluatedKey
func (m *Dynamodb) FindAllTasksByState(state, cursor string, asc bool, size int64) (taskStates []*TaskWithSignature, next string, err error) {
	if size <= 0 {
		size = 10
	}

	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(m.cnf.DynamoDB.TaskStatesTable),
		IndexName:              aws.String(tasks.TaskStateIndex), // use secondary global index
		Limit:                  aws.Int64(size),
		ProjectionExpression:   aws.String("TaskUUID, #st, TaskName, #err, Signature, CreatedAt"),
		KeyConditionExpression: aws.String("#st = :st"),
		ScanIndexForward:       aws.Bool(asc),
		ExpressionAttributeNames: map[string]*string{
			"#st":  aws.String("State"),
			"#err": aws.String("Error"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":st": {
				S: aws.String(state),
			},
		},
	}

	var lastEvaluatedKey map[string]*dynamodb.AttributeValue
	if cursor != "" {
		lastEvaluatedKey, err = decodeB64LastEvaluatedKey(cursor)
		if err != nil {
			log.ERROR.Println(err)
			return nil, next, err
		}
	}

	queryInput.ExclusiveStartKey = lastEvaluatedKey
	out, err := m.client.Query(queryInput)
	if err != nil {
		log.ERROR.Print(err)
		return nil, next, err
	}

	if out == nil {
		return nil, "", nil
	}

	if out.LastEvaluatedKey != nil {
		next, err = encodeB64LastEvaluatedKey(out.LastEvaluatedKey)
		if err != nil {
			log.ERROR.Println(err)
			return nil, next, err
		}
	}

	err = dynamodbattribute.UnmarshalListOfMaps(out.Items, &taskStates)
	if err != nil {
		log.ERROR.Print(err)
		return nil, next, err
	}

	return
}

// RerunTask :nodo:
func (m *Dynamodb) RerunTask(sig *tasks.Signature) error {
	sig.ETA = nil // reset ETA
	_, err := m.server.SendTask(sig)
	if err != nil {
		err = fmt.Errorf("failed to send task: %w", err)
		log.ERROR.Print(err)
		return err
	}
	return err
}

func decodeB64LastEvaluatedKey(cursor string) (key map[string]*dynamodb.AttributeValue, err error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cursor: %w", err)
	}

	err = json.Unmarshal(decoded, &key)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal decoded cursor: %s: %w", decoded, err)
	}

	return
}

func encodeB64LastEvaluatedKey(key map[string]*dynamodb.AttributeValue) (decoded string, err error) {
	bt, err := json.Marshal(key)
	if err != nil {
		return "", fmt.Errorf("failed to marshal LastEvaluatedKey :%w", err)
	}

	decoded = base64.StdEncoding.EncodeToString(bt)
	return
}
