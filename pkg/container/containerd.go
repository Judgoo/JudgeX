package main

import (
	"context"
	"fmt"
	"log"
	"syscall"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
)

func main() {
	if err := redisExample(); err != nil {
		log.Fatal(err)
	}
}

func redisExample() error {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return err
	}
	defer client.Close()
	ctx := namespaces.WithNamespace(context.Background(), "example")
	image, err := client.Pull(ctx, "docker.io/library/judgoo/gpp:v0.0.2", containerd.WithPullUnpack)
	if err != nil {
		return err
	}
	container, err := client.NewContainer(
		ctx,
		"redis-server",
		containerd.WithNewSnapshot("redis-server-snapshot", image),
		containerd.WithNewSpec(oci.WithImageConfig(image)),
	)
	if err != nil {
		return err
	}
	defer container.Delete(ctx, containerd.WithSnapshotCleanup)
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return err
	}
	defer task.Delete(ctx)
	exitStatusC, err := task.Wait(ctx)
	if err != nil {
		return err
	}

	if err := task.Start(ctx); err != nil {
		return err
	}
	time.Sleep(3 * time.Second)

	if err := task.Kill(ctx, syscall.SIGTERM); err != nil {
		return err
	}

	status := <-exitStatusC
	code, exitedAt, err := status.Result()
	if err != nil {
		return err
	}
	fmt.Printf("redis-server exited with status: %d exitedAt: %s\n", code, exitedAt)
	return nil
}
