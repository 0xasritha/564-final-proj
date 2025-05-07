package task

import (
	"os/exec"
)

type ShellCommand struct {
	ID              uint `json:"id"`
	commandWithArgs []string
}

func NewShellCommand(ID uint, commandWithArgs []string) *ShellCommand {
	return &ShellCommand{
		ID:              ID,
		commandWithArgs: commandWithArgs,
	}
}
func (s *ShellCommand) GetID() uint {
	return s.ID
}

func (s *ShellCommand) Do() Result {

	cmd := exec.Command(s.commandWithArgs[0], s.commandWithArgs[1:]...)

	output, err := cmd.CombinedOutput() //  CombinedOutput runs the command and returns its combined standard output and standard error.

	var result Result

	if err != nil {
		result.Success = false
		result.Content = err.Error()
	} else {
		result.Success = true
		result.Content = string(output)
	}

	return result
}
