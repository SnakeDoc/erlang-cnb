package erlang_test

import (
	"os"
	"strings"
	"testing"

	"github.com/SnakeDoc/erlang-cnb"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testErlangVersionResolver(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("ParseLatestStableVersion", func() {
		it("finds the latest stable version", func() {
			input := `	OTP-27.2 hash1 2024-12-11T10:30:23Z
					OTP-26.2.5.4 hash2 2024-10-09T10:56:17Z
					OTP-28.1.1 hash3 2025-10-20T15:23:31Z
					OTP-28.0 hash4 2025-05-21T08:33:49Z
					OTP-28.1 hash5 2025-09-17T08:23:36Z
				`

			result, err := erlang.ParseLatestStableVersion(strings.NewReader(input))
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("OTP-28.1.1"))
		})

		it("finds semver 2-part stable versions", func() {
			input := `	OTP-26.1 hash1 2024-04-25T23:10:07Z
					OTP-26.2 hash2 2024-04-25T22:52:52Z
					OTP-22.2.1 hash3 2024-04-25T22:40:28Z
					OTP-23.2.2 hash4 2024-04-25T22:40:25Z
					OTP-24.2.5.1 hash5 2024-06-25T09:44:43Z
					OTP-25.2.5.10 hash6 2025-03-28T15:26:22Z
				`

			result, err := erlang.ParseLatestStableVersion(strings.NewReader(input))
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("OTP-26.2"))
		})

		it("finds semver 3-part stable versions", func() {
			input := `	OTP-26.1 hash1 2024-04-25T23:10:07Z
					OTP-26.2 hash2 2024-04-25T22:52:52Z
					OTP-26.2.1 hash3 2024-04-25T22:40:28Z
					OTP-26.2.2 hash4 2024-04-25T22:40:25Z
				`

			result, err := erlang.ParseLatestStableVersion(strings.NewReader(input))
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("OTP-26.2.2"))
		})

		it("finds semver 4-part stable versions", func() {
			input := `	OTP-26.1 hash1 2024-04-25T23:10:07Z
					OTP-26.2 hash2 2024-04-25T22:52:52Z
					OTP-26.2.1 hash3 2024-04-25T22:40:28Z
					OTP-26.2.2 hash4 2024-04-25T22:40:25Z
					OTP-26.2.5.1 hash5 2024-06-25T09:44:43Z
					OTP-26.2.5.10 hash6 2025-03-28T15:26:22Z
				`

			result, err := erlang.ParseLatestStableVersion(strings.NewReader(input))
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("OTP-26.2.5.10"))
		})

		it("filters out pre-release versions", func() {
			input := `	OTP-27.0 hash1 2024-05-20T09:54:18Z
					OTP-27.0-rc1 hash2 2024-04-25T22:11:25Z
					OTP-27.0-rc2 hash3 2024-04-25T22:11:21Z
					OTP-27.0-rc3 hash4 2024-04-25T22:10:15Z
				`
			result, err := erlang.ParseLatestStableVersion(strings.NewReader(input))
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("OTP-27.0"))
		})

		it("filters out maintenance versions", func() {
			input := `	OTP-28.0.3 hash1 2025-09-10T15:26:54Z
					maint hash2 2025-10-31T16:50:31Z
					maint-27 hash3 2025-10-28T08:27:27Z
					maint-28 hash4 2025-10-20T15:10:55Z
				`
			result, err := erlang.ParseLatestStableVersion(strings.NewReader(input))
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("OTP-28.0.3"))
		})

		it("filters out master versions", func() {
			input := `	OTP-26.2.5.15 hash1 2025-09-10T16:19:34Z
					master hash2 2025-10-31T16:50:31Z
				`
			result, err := erlang.ParseLatestStableVersion(strings.NewReader(input))
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("OTP-26.2.5.15"))
		})

		context("when no stable versions are found", func() {
			it("returns an error", func() {
				input := `	OTP-28.0-rc1 hash1 2025-02-12T11:05:16Z
						maint hash2 2025-10-30T17:20:29Z
						master hash3 2025-10-30T17:21:08Z
					`
				_, err := erlang.ParseLatestStableVersion(strings.NewReader(input))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no stable Erlang versions found"))
			})
		})
	})

	context("ResolveVersion", func() {
		context("when BP_ERLANG_VERSION is set", func() {
			it.Before(func() {
				os.Setenv("BP_ERLANG_VERSION", "28.1.1")
			})
			it.After(func() {
				os.Unsetenv("BP_ERLANG_VERSION")
			})

			it("returns the normalized version from the environment variable", func() {
				result, err := erlang.ResolveVersion("amd64", "ubuntu-24.04")
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal("28.1.1"))
			})
		})
	})

}
