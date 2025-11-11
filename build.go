package erlang

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

const (
	Erlang           = "erlang"
	LayerName        = "erlang"
	VersionKey       = "version"
	ArchKey          = "arch"
	UbuntuVersionKey = "ubuntu-version"
)

//go:generate faux --interface Installer --output fakes/installer.go
type Installer interface {
	BuildDownloadURL(arch, ubuntuVersion, version string) string
	Install(url, layerPath string) error
}

func Build(installer Installer, logger scribe.Emitter, clock chronos.Clock) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		// detect architecture and ubuntu version
		arch := runtime.GOARCH
		ubuntuVersion, err := detectUbuntuVersion(context.Stack)
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to detect ubuntu version: %w", err)
		}

		logger.Process("Resolving Erlang version")
		logger.Subprocess("Architecture: %s", arch)
		logger.Subprocess("Stack: %s (%s)", context.Stack, ubuntuVersion)

		// resolve which version to install
		version, err := ResolveVersion(arch, ubuntuVersion)
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to resolve Erlang version: %w", err)
		}

		logger.Action("Using Erlang version: %s", version)
		logger.Break()

		// get or create the erlang layer
		erlangLayer, err := context.Layers.Get(LayerName)
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to get %s layer: %w", LayerName, err)
		}

		// check if we can reuse the existing layer
		cachedVersion, _ := erlangLayer.Metadata[VersionKey].(string)
		cachedArch, _ := erlangLayer.Metadata[ArchKey].(string)
		cachedUbuntuVersion, _ := erlangLayer.Metadata[UbuntuVersionKey].(string)

		if cachedVersion == version && cachedArch == arch && cachedUbuntuVersion == ubuntuVersion {
			logger.Process("Reusing cached layer %s", erlangLayer.Path)
			logger.Break()

			erlangLayer.Launch = true
			erlangLayer.Cache = true
			erlangLayer.Build = true

			return packit.BuildResult{
				Layers: []packit.Layer{erlangLayer},
			}, nil
		}

		// need to install - reset the layer
		logger.Process("Executing build process")

		erlangLayer, err = erlangLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to reset %s layer: %w", LayerName, err)
		}

		// install erlang
		downloadURL := installer.BuildDownloadURL(arch, ubuntuVersion, version)

		logger.Subprocess("Downloading Erlang %s", version)
		logger.Action("Source: %s", downloadURL)

		duration, err := clock.Measure(func() error {
			return installer.Install(downloadURL, erlangLayer.Path)
		})
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to install Erlang %s: %w", version, err)
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond).String())
		logger.Break()

		// setup env variables
		erlangLayer.SharedEnv.Default("ERLANG_HOME", erlangLayer.Path)
		erlangLayer.SharedEnv.Prepend("PATH", filepath.Join(erlangLayer.Path, "bin"), ":")

		// store metadata for caching
		erlangLayer.Metadata = map[string]any{
			VersionKey:       version,
			ArchKey:          arch,
			UbuntuVersionKey: ubuntuVersion,
		}

		erlangLayer.Launch = true
		erlangLayer.Cache = true
		erlangLayer.Build = true

		logger.EnvironmentVariables(erlangLayer)

		return packit.BuildResult{
			Layers: []packit.Layer{erlangLayer},
		}, nil
	}
}

func detectUbuntuVersion(stackID string) (string, error) {
	switch {
	case strings.Contains(stackID, "noble"):
		return "ubuntu-24.04", nil
	case strings.Contains(stackID, "jammy"):
		return "ubuntu-22.04", nil
	case strings.Contains(stackID, "focal"):
		return "ubuntu-20.04", nil
	case strings.Contains(stackID, "bionic"):
		return "ubuntu-18.04", nil
	default:
		return "", fmt.Errorf("unsupported stack: %s - this buildpack only supports Ubuntu-based stacks (noble, jammy, focal, bionic)", stackID)
	}
}
