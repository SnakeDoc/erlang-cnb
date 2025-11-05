package erlang_test

import (
	"archive/tar"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/SnakeDoc/erlang-cnb"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testErlangInstaller(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("BuildDownloadURL", func() {
		it("constructs the correct download URL", func() {
			installer := erlang.NewErlangInstaller()

			url := installer.BuildDownloadURL("amd64", "ubuntu-24.04", "OTP-28.1.1")
			Expect(url).To(Equal("https://builds.hex.pm/builds/otp/amd64/ubuntu-24.04/OTP-28.1.1.tar.gz"))
		})

		it("works with amd64 architecture", func() {
			installer := erlang.NewErlangInstaller()

			url := installer.BuildDownloadURL("amd64", "ubuntu-20.04", "OTP-24.0.5")
			Expect(url).To(Equal("https://builds.hex.pm/builds/otp/amd64/ubuntu-20.04/OTP-24.0.5.tar.gz"))
		})

		it("works with arm64 architecture", func() {
			installer := erlang.NewErlangInstaller()

			url := installer.BuildDownloadURL("arm64", "ubuntu-22.04", "OTP-27.3.4")
			Expect(url).To(Equal("https://builds.hex.pm/builds/otp/arm64/ubuntu-22.04/OTP-27.3.4.tar.gz"))
		})
	})

	context("Install", func() {
		var (
			installer erlang.ErlangInstaller
			layerPath string
			server    *httptest.Server
		)

		it.Before(func() {
			installer = erlang.NewErlangInstaller()

			var err error
			layerPath, err = os.MkdirTemp("", "layer")
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			if server != nil {
				server.Close()
			}
			Expect(os.RemoveAll(layerPath)).To(Succeed())
		})

		it("downloads and extracts Erlang", func() {
			// create a test tarball with a nested directory structure
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gw := gzip.NewWriter(w)
				defer gw.Close()

				tw := tar.NewWriter(gw)
				defer tw.Close()

				// simulate strip-components=1 by having content in a subdirectory header for directory
				tw.WriteHeader(&tar.Header{
					Name:     "otp-28.1.1/",
					Mode:     0755,
					Typeflag: tar.TypeDir,
				})

				content := []byte("test erlang binary")
				tw.WriteHeader(&tar.Header{
					Name: "otp-28.1.1/bin/erl",
					Mode: 0755,
					Size: int64(len(content)),
				})
				tw.Write(content)
			}))

			err := installer.Install(server.URL, layerPath)
			Expect(err).NotTo(HaveOccurred())

			erlPath := filepath.Join(layerPath, "bin", "erl")
			Expect(erlPath).To(BeARegularFile())

			content, err := os.ReadFile(erlPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("test erlang binary"))
		})

		context("failure cases", func() {
			context("when the download fails", func() {
				it("returns an error", func() {
					// create a server and close it immediately to simulate failure
					server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
					url := server.URL
					server.Close()

					err := installer.Install(url, layerPath)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to download Erlang"))
				})
			})

			context("when the server returns non-200 status", func() {
				it("returns an error", func() {
					server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
					}))

					err := installer.Install(server.URL, layerPath)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("received status code 404"))
				})
			})

			context("when the archive is invalid", func() {
				it("returns an error", func() {
					server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "application/gzip")
						// incomplete gzip header - definitely will fail to decompress
						// 0x1f, 0x8b - indicates gzip format
						// missing rest of gzip header and data to ensure failure
						w.Write([]byte{0x1f, 0x8b})
						w.Write([]byte("invalid data"))
					}))

					err := installer.Install(server.URL, layerPath)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to decompress"))
				})
			})
		})
	})
}
