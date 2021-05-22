package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-cmd/cmd"
)

func Exec(str string, dir string) *cmd.Status {
	target := strings.Split(str, " ")

	dockerCmd := cmd.NewCmd(target[0], target[1:]...)
	if dir != "" {
		dockerCmd.Dir = dir
	}
	statusChan := dockerCmd.Start() // non-blocking

	stopCh := make(chan struct{})
	// 2 分钟后杀死进程
	go func() {
		t := time.After(2 * time.Minute)
		for {
			// Check if command is done
			select {
			case <-stopCh:
				fmt.Println("done cmd")
				t = nil
				return
			case <-t:
				fmt.Println("stop docker cmd")
				dockerCmd.Stop()
				return
			}
		}
	}()

	// Block waiting for command to exit, be stopped, or be killed
	finalStatus := <-statusChan
	close(stopCh)
	return &finalStatus
}
