package erlang

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/paketo-buildpacks/packit/v2/vacation"
)

const (
	DownloadURLTemplate = "https://builds.hex.pm/builds/otp/%s/%s/%s.tar.gz"
)

type ErlangInstaller struct{}

func NewErlangInstaller() ErlangInstaller {
	return ErlangInstaller{}
}

func formatOTPVersion(version string) string {
	if !strings.HasPrefix(version, "OTP-") {
		return "OTP-" + version
	}
	return version
}

func (i ErlangInstaller) BuildDownloadURL(arch, ubuntuVersion, version string) string {
	normalizedVersion := formatOTPVersion(version)
	return fmt.Sprintf(DownloadURLTemplate, arch, ubuntuVersion, normalizedVersion)
}

func (i ErlangInstaller) Install(url, layerPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download Erlang from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download Erlang from %s: received status code %d", url, resp.StatusCode)
	}

	err = vacation.NewArchive(resp.Body).StripComponents(1).Decompress(layerPath)
	if err != nil {
		return fmt.Errorf("failed to decompress Erlang archive to %s: %w", layerPath, err)
	}

	return nil
}
