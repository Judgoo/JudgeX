package utils

import (
	"io"
	"strings"
	"time"

	"github.com/Judgoo/JudgeX/logger"
	"github.com/go-cmd/cmd"
)

func Exec(cmdStr string, dir string, stdin io.Reader) *cmd.Status {
	target := strings.Fields(cmdStr)
	logger.Sugar.Infof("run cmd %s", cmdStr)

	dockerCmd := cmd.NewCmd(target[0], target[1:]...)
	if dir != "" {
		dockerCmd.Dir = dir
	}
	var statusChan <-chan cmd.Status
	if stdin != nil {
		statusChan = dockerCmd.StartWithStdin(stdin)
	} else {
		statusChan = dockerCmd.Start() // non-blocking
	}

	stopCh := make(chan struct{})
	// 2 分钟后杀死进程
	go func() {
		t := time.After(2 * time.Minute)
		// Check if command is done
		select {
		case <-stopCh:
			logger.Sugar.Info("done cmd")
			t = nil
			return
		case <-t:
			logger.Sugar.Info("stop docker cmd")
			dockerCmd.Stop()
			return
		}
	}()
	// Block waiting for command to exit, be stopped, or be killed
	finalStatus := <-statusChan
	close(stopCh)
	return &finalStatus
}
