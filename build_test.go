package erlang_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/SnakeDoc/erlang-cnb"
	"github.com/SnakeDoc/erlang-cnb/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir  string
		workingDir string
		cnbDir     string
		buffer     *bytes.Buffer
		timeStamp  time.Time
		installer  *fakes.Installer

		build        packit.BuildFunc
		buildContext packit.BuildContext
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.Setenv("BP_ERLANG_VERSION", "28.1.1")).To(Succeed())

		buffer = bytes.NewBuffer(nil)
		timeStamp = time.Now()
		installer = &fakes.Installer{}

		build = erlang.Build(
			installer,
			scribe.NewEmitter(buffer),
			chronos.NewClock(func() time.Time { return timeStamp }),
		)

		buildContext = packit.BuildContext{
			CNBPath: cnbDir,
			Stack:   "io.buildpacks.stacks.jammy",
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Erlang Buildpack",
				Version: "0.0.1",
			},
			Layers:     packit.Layers{Path: layersDir},
			WorkingDir: workingDir,
		}
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
		Expect(os.Unsetenv("BP_ERLANG_VERSION")).To(Succeed())
	})

	it("returns a result that installs erlang", func() {
		result, err := build(buildContext)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		layer := result.Layers[0]

		Expect(layer.Name).To(Equal("erlang"))
		Expect(layer.Path).To(Equal(filepath.Join(layersDir, "erlang")))

		Expect(layer.SharedEnv).To(HaveKeyWithValue("ERLANG_HOME.default", filepath.Join(layersDir, "erlang")))
		Expect(layer.SharedEnv).To(HaveKeyWithValue("PATH.prepend", filepath.Join(layersDir, "erlang", "bin")))
		Expect(layer.SharedEnv).To(HaveKeyWithValue("PATH.delim", ":"))

		Expect(layer.Build).To(BeTrue())
		Expect(layer.Launch).To(BeTrue())
		Expect(layer.Cache).To(BeTrue())

		Expect(layer.Metadata).To(HaveKeyWithValue("version", "28.1.1"))
		Expect(layer.Metadata).To(HaveKeyWithValue("arch", "amd64"))
		Expect(layer.Metadata).To(HaveKeyWithValue("ubuntu-version", "ubuntu-22.04"))

		Expect(buffer.String()).To(ContainSubstring("Some Erlang Buildpack 0.0.1"))
		Expect(buffer.String()).To(ContainSubstring("Resolving Erlang version"))
		Expect(buffer.String()).To(ContainSubstring("Architecture: amd64"))
		Expect(buffer.String()).To(ContainSubstring("Stack: io.buildpacks.stacks.jammy (ubuntu-22.04)"))
		Expect(buffer.String()).To(ContainSubstring("Using Erlang version: 28.1.1"))
		Expect(buffer.String()).To(ContainSubstring("Executing build process"))
		Expect(buffer.String()).To(ContainSubstring("Downloading Erlang 28.1.1"))
	})

	context("when the layer is already cached", func() {
		it.Before(func() {
			err := os.MkdirAll(filepath.Join(layersDir, "erlang"), 0755)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(layersDir, "erlang.toml"), []byte(`
				[metadata]
				version = "28.1.1"
				arch = "amd64"
				ubuntu-version = "ubuntu-22.04"
			`), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		it("reuses the cached layer", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Build).To(BeTrue())
			Expect(layer.Launch).To(BeTrue())
			Expect(layer.Cache).To(BeTrue())

			Expect(buffer.String()).To(ContainSubstring("Reusing cached layer"))
			Expect(buffer.String()).NotTo(ContainSubstring("Downloading Erlang"))
		})
	})

	context("when the cached layer has a different version", func() {
		it.Before(func() {
			err := os.MkdirAll(filepath.Join(layersDir, "erlang"), 0755)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(layersDir, "erlang.toml"), []byte(`
				[metadata]
				version = "27.3.4"
				arch = "amd64"
				ubuntu-version = "ubuntu-22.04"
			`), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		it("rebuilds the layer", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Metadata).To(HaveKeyWithValue("version", "28.1.1"))

			Expect(buffer.String()).NotTo(ContainSubstring("Reusing cached layer"))
			Expect(buffer.String()).To(ContainSubstring("Executing build process"))
			Expect(buffer.String()).To(ContainSubstring("Downloading Erlang 28.1.1"))
		})
	})

	context("when using different stacks", func() {
		it("maps noble to ubuntu-24.04", func() {
			buildContext.Stack = "io.buildpacks.stacks.noble"

			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			layer := result.Layers[0]
			Expect(layer.Metadata).To(HaveKeyWithValue("ubuntu-version", "ubuntu-24.04"))

			Expect(buffer.String()).To(ContainSubstring("ubuntu-24.04"))
		})

		it("maps focal to ubuntu-20.04", func() {
			buildContext.Stack = "io.buildpacks.stacks.focal"

			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			layer := result.Layers[0]
			Expect(layer.Metadata).To(HaveKeyWithValue("ubuntu-version", "ubuntu-20.04"))

			Expect(buffer.String()).To(ContainSubstring("ubuntu-20.04"))
		})
	})

	context("failure cases", func() {
		context("when the stack is unsupported", func() {
			it("returns an error", func() {
				buildContext.Stack = "io.buildpacks.stacks.alpine"

				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("unsupported stack")))
				Expect(err).To(MatchError(ContainSubstring("alpine")))
			})
		})

		context("when version resolution fails", func() {
			it.Before(func() {
				// unset the env var to trigger network resolution
				Expect(os.Unsetenv("BP_ERLANG_VERSION")).To(Succeed())

				// use an invalid stack to cause detectUbuntuVersion to fail
				buildContext.Stack = "invalid-stack"
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("unsupported stack")))
			})
		})

		context("when the layer cannot be retrieved", func() {
			it.Before(func() {
				Expect(os.Chmod(layersDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(layersDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
	})
}
