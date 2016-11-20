package bond

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type BuildInfo struct {
	Version string
	Options BuildOptions
}

func GetInfoFromFileName(fileName string) (BuildInfo, error) {
	info := BuildInfo{Options: BuildOptions{}}
	fileName = filepath.Base(fileName)

	if strings.Contains(fileName, "debugsymbols") {
		info.Options.Debug = true
	}

	// TODO probably need to make this a string.
	for _, arch := range []MongoDBArch{AMD64, X86, POWER, ZSeries} {
		if strings.Contains(fileName, string(arch)) {
			info.Options.Arch = arch
			break
		}
	}

	if info.Options.Arch == "" {
		return BuildInfo{}, errors.Errorf("path '%s' does not  ")
	}

	edition, err := getEdition(fileName)
	if err != nil {
		return BuildInfo{}, errors.Wrap(err, "problem resolving edition")
	}
	info.Options.Edition = edition

	target, err := getTarget(fileName)
	if err != nil {
		return BuildInfo{}, errors.Wrap(err, "problem resolving target")
	}
	info.Options.Target = target

	if err != nil {
		return BuildInfo{}, errors.Wrap(err, "problem resolving version")
	}

	return info, nil
}

func getVersion(fn string) (string, error) {
	parts := strings.Split(fn, "-")
	if len(parts) <= 2 {
		return "", errors.Errorf("path %s does not contain enough elements to include a version", fn)
	}

	isNightly := strings.Contains(fn, "~")
	isRc := strings.Contains(parts[len(parts)-1], "rc")

	var rIdx int

	// MUST WRITE TESTS FOR THIS

	if isRc {
		if isNightly {
			rIdx = len(parts) - 3
		} else {
			rIdx = len(parts) - 2
		}
	} else {
		if isNightly {
			rIdx = len(parts) - 1
		} else {
			rIdx = len(parts)
		}
	}

	if rIdx <= 2 {
		return "", errors.Errorf("%s is an invalid file name", fn)
	}

	return strings.Join(parts[rIdx-1:], "-"), nil

}

func getEdition(fn string) (MongoDBEdition, error) {
	if strings.Contains(fn, "enterprise") {
		return Enterprise, nil
	}

	for _, distro := range []string{"rhel", "suse", "2008", "osx-ssl", "debian", "ubuntu", "amazon"} {
		if strings.Contains(fn, distro) {
			return CommunityTargeted, nil
		}
	}

	for _, platform := range []string{"osx", "win32", "sunos5", "Linux"} {
		if strings.HasPrefix(fn, "manged-"+platform) {
			return Base, nil
		}
	}

	return "", errors.Errorf("path %s does not have a valid edition", fn)
}

func getTarget(fn string) (string, error) {
	// enterprise targets:
	if strings.Contains(fn, "enterprise") {
		for _, platform := range []string{"osx", "windows"} {
			if strings.Contains(fn, platform) {
				return platform, nil
			}
		}
		if strings.Contains(fn, "linux") {
			return strings.Split(fn, "-")[3], nil
		}
	}

	// all base and community targeted cases

	// OSX variants
	if strings.Contains(fn, "osx-ssl") {
		return "osx-ssl", nil
	}
	if strings.Contains(fn, "osx") {
		return "osx", nil
	}

	// all windows windows
	if strings.Contains(fn, "2008plus-ssl") {
		return "windows_x86_64-2008plus-ssl", nil
	}
	if strings.Contains(fn, "2008plus") {
		return "windows_x86_64-2008plus", nil
	}
	if strings.Contains(fn, "win32-i386") {
		return "windows_i686", nil
	}
	if strings.Contains(fn, "win32-x86_64") {
		return "windows_x86_64", nil
	}

	// linux base distro
	if strings.Contains(fn, "linux-x86_64") {
		return "linux_x86_64", nil
	}
	if strings.Contains(fn, "linux-i386") {
		return "linux_i386", nil
	}

	// solaris!
	if strings.Contains(fn, "sunos5") {
		return "sunos5", nil
	}

	return "", errors.Errorf("could not determine platform for %s", fn)
}
