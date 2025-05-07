package c2communicator

import (
	"asritha.dev/implant/pkg/task"
)

type C2Communicator interface {
	Beacon(implantID uint, results []task.Result) (*task.Queue, error)
}
