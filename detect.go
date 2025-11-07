package erlang

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
)

//go:generate faux --interface VersionParser --output fakes/version_parser.go
type VersionParser interface {
	ParseVersion(path string) (version string, err error)
}

type BuildPlanMetadata struct {
	Version       string `toml:"version"`
	VersionSource string `toml:"version-source"`
}

func Detect(toolVersionsParser VersionParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		var requirements []packit.BuildPlanRequirement

		version := os.Getenv("BP_ERLANG_VERSION")
		if version != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "erlang",
				Metadata: BuildPlanMetadata{
					Version:       version,
					VersionSource: "BP_ERLANG_VERSION",
				},
			})
		}

		version, err := toolVersionsParser.ParseVersion(filepath.Join(context.WorkingDir, ".tool-versions"))
		if err != nil {
			return packit.DetectResult{}, err
		}

		if version != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "erlang",
				Metadata: BuildPlanMetadata{
					Version:       version,
					VersionSource: ".tool-versions",
				},
			})
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "erlang"},
				},
				Requires: requirements,
			},
		}, nil
	}
}
