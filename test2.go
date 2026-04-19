package main
import (
	"context"
	"github.com/moby/moby/client"
)
func main() {
	ctx := context.Background()
	cli, _ := client.NewClientWithOpts()
	_,_ = cli.ImagePull(ctx, "a", client.ImagePullOptions{})
	res, _ := cli.ContainerCreate(ctx, client.ContainerCreateOptions{Image:"a"})
	_, _ = cli.ContainerStart(ctx, res.ID, client.ContainerStartOptions{})
}
