package container

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/Judgoo/JudgeX/pkg/entities"
	"github.com/containers/podman/v3/libpod/define"
	"github.com/containers/podman/v3/pkg/bindings"
	"github.com/containers/podman/v3/pkg/bindings/containers"
	"github.com/containers/podman/v3/pkg/bindings/images"
	"github.com/containers/podman/v3/pkg/specgen"
)

type Podman struct {
	ctx *context.Context
}

func New() *Podman {
	fmt.Println("Welcome to the Podman Go bindings tutorial")

	// Get Podman socket location
	sock_dir := os.Getenv("XDG_RUNTIME_DIR")
	socket := "unix:" + sock_dir + "/podman/podman.sock"

	// Connect to Podman socket
	connText, err := bindings.NewConnection(context.Background(), socket)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return &Podman{
		ctx: &connText,
	}
}
func BoolAddr(b bool) *bool {
	boolVar := b
	return &boolVar
}
func Int64Addr(b int64) *int64 {
	_var := b
	return &_var
}
func (p *Podman) Run(imageName string, dataBytes []byte) (*entities.JudgerResult, error) {
	s := specgen.NewSpecGenerator(imageName, false)
	s.Terminal = true
	s.Privileged = true
	s.Stdin = true
	// s.Name = "JudgeX"
	// s.N
	// s.ResourceLimits.CPU.Cpus = "2"
	// s.ResourceLimits.Memory.Limit = Int64Addr(100 * 1024 * 1024)
	r, err := containers.CreateWithSpec(*p.ctx, s, &containers.CreateOptions{})
	if err != nil {
		fmt.Println(err)
		return &entities.JudgerResult{}, err
	}

	stdin := bytes.NewReader(dataBytes)
	stdoutBuffer := new(bytes.Buffer)
	stderrBuffer := new(bytes.Buffer)

	attachReady := make(chan bool)

	// Container start
	fmt.Println("Starting container...")
	err = containers.Start(*p.ctx, r.ID, nil)
	if err != nil {
		fmt.Println(err)
		return &entities.JudgerResult{}, err
	}

	fmt.Println("Attaching container...")
	err = containers.Attach(*p.ctx, r.ID, stdin, stdoutBuffer, stderrBuffer, attachReady, &containers.AttachOptions{
		Logs:   BoolAddr(true),
		Stream: BoolAddr(true),
	})
	if err != nil {
		fmt.Println(err)
		return &entities.JudgerResult{}, err
	}
	// <-attachReady

	fmt.Println("Wait container...")
	exitcode, eerr := containers.Wait(*p.ctx, r.ID, &containers.WaitOptions{
		Condition: []define.ContainerStatus{define.ContainerStateRunning},
	})

	if eerr != nil {
		fmt.Println(eerr)
		return &entities.JudgerResult{}, eerr
	}
	return &entities.JudgerResult{
		Stdout: stdoutBuffer.String(),
		Stderr: stderrBuffer.String(),
		Exit:   exitcode,
	}, nil
}

func (p *Podman) Pull() {
	// Pull Busybox image (Sample 1)
	fmt.Println("Pulling Busybox image...")
	_, err := images.Pull(*p.ctx, "docker.io/busybox", &images.PullOptions{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Pull Fedora image (Sample 2)
	rawImage := "registry.fedoraproject.org/fedora:latest"
	fmt.Println("Pulling Fedora image...")
	_, err = images.Pull(*p.ctx, rawImage, &images.PullOptions{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (p *Podman) List() {
	// List images
	imageSummary, err := images.List(*p.ctx, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var names []string
	for _, i := range imageSummary {
		names = append(names, i.RepoTags...)
	}
	fmt.Println("Listing images...")
	fmt.Println(names)
}
