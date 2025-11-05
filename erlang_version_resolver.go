package erlang

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	BuildsURLTemplate = "https://builds.hex.pm/builds/otp/%s/%s/builds.txt"
)

func ResolveVersion(arch, ubuntuVersion string) (string, error) {
	if version := os.Getenv("BP_ERLANG_VERSION"); version != "" {
		return NormalizeVersion(version), nil
	}

	return fetchLatestVersion(arch, ubuntuVersion)
}

func NormalizeVersion(version string) string {
	if !strings.HasPrefix(version, "OTP-") {
		return "OTP-" + version
	}
	return version
}

func fetchLatestVersion(arch, ubuntuVersion string) (string, error) {
	url := fmt.Sprintf(BuildsURLTemplate, arch, ubuntuVersion)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest Erlang version from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch latest Erlang version from %s: received status code %d", url, resp.StatusCode)
	}

	return ParseLatestStableVersion(resp.Body)
}

// stable releases have the format "OTP-x.y.z.*"
func ParseLatestStableVersion(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	var latestVer []int
	var latest string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		versionTag := fields[0]

		if !isStableVersion(versionTag) {
			continue
		}

		versionPart := strings.TrimPrefix(versionTag, "OTP-")
		ver, _ := parseVersion(versionPart)

		if latestVer == nil || compareVersions(ver, latestVer) > 0 {
			latestVer = ver
			latest = versionTag
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading version data: %w", err)
	}

	if latest == "" {
		return "", fmt.Errorf("no stable Erlang versions found")
	}

	return latest, nil
}

func isStableVersion(versionTag string) bool {
	if !strings.HasPrefix(versionTag, "OTP-") {
		return false
	}

	versionPart := strings.TrimPrefix(versionTag, "OTP-")

	_, err := parseVersion(versionPart)
	if err != nil {
		return false
	}

	if hasPrerelease(versionPart) {
		return false
	}

	return true
}

func parseVersion(version string) ([]int, error) {
	parts := strings.Split(version, ".")
	result := make([]int, 0, len(parts))

	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid version part: %s", part)
		}
		result = append(result, num)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("empty version")
	}

	return result, nil
}

func hasPrerelease(version string) bool {
	for _, ch := range version {
		if ch != '.' && (ch < '0' || ch > '9') {
			return true
		}
	}
	return false
}

func compareVersions(v1, v2 []int) int {
	maxLen := max(len(v2), len(v1))

	for i := range maxLen {
		var n1, n2 int

		if i < len(v1) {
			n1 = v1[i]
		}
		if i < len(v2) {
			n2 = v2[i]
		}

		if n1 != n2 {
			return n1 - n2
		}
	}
	return 0
}
