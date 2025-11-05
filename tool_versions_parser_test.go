package erlang_test

import (
	"os"
	"testing"

	"github.com/SnakeDoc/erlang-cnb"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testToolVersionsParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path   string
		parser erlang.ToolVersionsParser
	)

	it.Before(func() {
		file, err := os.CreateTemp("", ".tool-versions")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		path = file.Name()
		parser = erlang.NewToolVersionsParser()
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	it("parses erlang version from .tool-versions", func() {
		err := os.WriteFile(path, []byte("erlang 28.1.1"), 0644)
		Expect(err).NotTo(HaveOccurred())

		version, err := parser.ParseVersion(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(version).To(Equal("OTP-28.1.1"))
	})

	it("finds erlang version among multiple tools", func() {
		content := `	nodejs 24.11.0
				erlang 28.1.1
				elixir 1.19.0
			`

		err := os.WriteFile(path, []byte(content), 0644)
		Expect(err).NotTo(HaveOccurred())

		version, err := parser.ParseVersion(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(version).To(Equal("OTP-28.1.1"))
	})

	it("normalizes version without OTP- prefix", func() {
		err := os.WriteFile(path, []byte("erlang 26.2.5"), 0644)
		Expect(err).NotTo(HaveOccurred())

		version, err := parser.ParseVersion(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(version).To(Equal("OTP-26.2.5"))
	})

	it("preserves OTP- prefix if present", func() {
		err := os.WriteFile(path, []byte("erlang OTP-27.3.4"), 0644)
		Expect(err).NotTo(HaveOccurred())

		version, err := parser.ParseVersion(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(version).To(Equal("OTP-27.3.4"))
	})

	it("skips comments and empty lines", func() {
		content := `	# This is a comment
				nodejs 24.11.0

				# Erlang version
				erlang 28.1.1
				elixir 1.19.0`

		err := os.WriteFile(path, []byte(content), 0644)
		Expect(err).NotTo(HaveOccurred())

		version, err := parser.ParseVersion(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(version).To(Equal("OTP-28.1.1"))
	})

	it("handles whitespace variations", func() {
		content := `  erlang   28.0.1  `
		err := os.WriteFile(path, []byte(content), 0644)
		Expect(err).NotTo(HaveOccurred())

		version, err := parser.ParseVersion(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(version).To(Equal("OTP-28.0.1"))
	})

	context("when erlang is not specified", func() {
		it("returns empty string", func() {
			content := `	nodejs 20.11.0
					elixir 1.15.7
				`

			err := os.WriteFile(path, []byte(content), 0644)
			Expect(err).NotTo(HaveOccurred())

			version, err := parser.ParseVersion(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(BeEmpty())
		})
	})

	context("when the .tool-versions file does not exist", func() {
		it.Before(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		it("returns empty string without error", func() {
			version, err := parser.ParseVersion(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(BeEmpty())
		})
	})

	context("when the file is empty", func() {
		it("returns empty string", func() {
			err := os.WriteFile(path, []byte(""), 0644)
			Expect(err).NotTo(HaveOccurred())

			version, err := parser.ParseVersion(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(BeEmpty())
		})
	})

	context("when the file has only comments", func() {
		it("returns empty string", func() {
			content := `	# Comment 1
					# Comment 2
					# Comment 3
				`

			err := os.WriteFile(path, []byte(content), 0644)
			Expect(err).NotTo(HaveOccurred())

			version, err := parser.ParseVersion(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(BeEmpty())
		})
	})
}
