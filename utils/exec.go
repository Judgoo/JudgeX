package utils

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

var ErrCommandInvalid = errors.New("error command string")

func Exec(str string, dir string) (string, string, int, error) {
	var (
		stdout, stderr bytes.Buffer
		err            error
	)
	target := strings.Split(str, " ")
	if len(target) < 1 {
		return "", "", 0, ErrCommandInvalid
	}
	runner := exec.Command(target[0], target[1:]...)
	if dir != "" {
		runner.Dir = dir
	}
	runner.Stdout = &stdout
	runner.Stderr = &stderr
	err = runner.Run()
	return stdout.String(), stderr.String(), runner.ProcessState.ExitCode(), err
}
