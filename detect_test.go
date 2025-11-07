package erlang_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SnakeDoc/erlang-cnb"
	"github.com/SnakeDoc/erlang-cnb/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir         string
		toolVersionsParser *fakes.VersionParser
		detect             packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		toolVersionsParser = &fakes.VersionParser{}

		detect = erlang.Detect(toolVersionsParser)
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
		Expect(os.Unsetenv("BP_ERLANG_VERSION")).To(Succeed())
	})

	it("always provides erlang", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: workingDir,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Plan).To(Equal(packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{
				{Name: "erlang"},
			},
		}))
	})

	context("when BP_ERLANG_VERSION is set", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_ERLANG_VERSION", "27.3.4")).To(Succeed())
		})

		it("requires erlang with version from environment variable", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "erlang"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "erlang",
						Metadata: erlang.BuildPlanMetadata{
							Version:       "27.3.4",
							VersionSource: "BP_ERLANG_VERSION",
						},
					},
				},
			}))
		})
	})

	context("when .tool-versions exists", func() {
		it.Before(func() {
			toolVersionsParser.ParseVersionCall.Returns.Version = "28.1.1"
		})

		it("requires erlang with version from .tool-versions", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "erlang"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "erlang",
						Metadata: erlang.BuildPlanMetadata{
							Version:       "28.1.1",
							VersionSource: ".tool-versions",
						},
					},
				},
			}))

			Expect(toolVersionsParser.ParseVersionCall.Receives.Path).To(Equal(filepath.Join(workingDir, ".tool-versions")))
		})
	})

	context("when both BP_ERLANG_VERSION and .tool-versions are set", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_ERLANG_VERSION", "27.3.4")).To(Succeed())
			toolVersionsParser.ParseVersionCall.Returns.Version = "28.1.1"
		})

		it("includes both requirements with BP_ERLANG_VERSION first", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "erlang"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "erlang",
						Metadata: erlang.BuildPlanMetadata{
							Version:       "27.3.4",
							VersionSource: "BP_ERLANG_VERSION",
						},
					},
					{
						Name: "erlang",
						Metadata: erlang.BuildPlanMetadata{
							Version:       "28.1.1",
							VersionSource: ".tool-versions",
						},
					},
				},
			}))
		})
	})

	context("failure cases", func() {
		context("when parsing .tool-versions fails", func() {
			it.Before(func() {
				toolVersionsParser.ParseVersionCall.Returns.Err = os.ErrPermission
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(os.ErrPermission))
			})
		})
	})
}
