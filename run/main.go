package main

import (
	"os"

	erlang "github.com/SnakeDoc/erlang-cnb"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

func main() {
	logger := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))

	packit.Run(
		erlang.Detect(),
		erlang.Build(logger),
	)
}
