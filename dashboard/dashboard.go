package dashboard

import (
	"github.com/RichardKnop/machinery/v1/tasks"
)

// Dashboard :noodc:
type Dashboard interface {
	FindAllTasksByState(state, cursor string, asc bool, size int64) (taskStates []*TaskWithSignature, next string, err error)
	RerunTask(sig *tasks.Signature) error
}
