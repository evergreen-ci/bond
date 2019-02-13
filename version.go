package bond

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// ArtifactVersion represents a document in the Version field of the
// MongoDB build information feed. See
// http://downloads.mongodb.org/full.json for an example.ownload
type ArtifactVersion struct {
	Version   string
	Downloads []ArtifactDownload
	GitHash   string

	ProductionRelease  bool `json:"production_release"`
	DevelopmentRelease bool `json:"development_release"`
	Current            bool

	table map[BuildOptions]ArtifactDownload
	mutex sync.RWMutex
}

func (version *ArtifactVersion) refresh() {
	version.mutex.Lock()
	defer version.mutex.Unlock()

	version.table = make(map[BuildOptions]ArtifactDownload)

	for _, dl := range version.Downloads {
		version.table[dl.GetBuildOptions()] = dl
	}
}

func parseVersionParts(version string) (major uint64, minor uint64, patch uint64, err error) {
	versionParts := strings.SplitN(version, ".", 3)
	if len(versionParts) != 3 {
		err = errors.New("version must be in the form major.minor.patch")
		return
	}
	major, err = strconv.ParseUint(versionParts[0], 10, 64)
	if err != nil {
		err = errors.Wrap(err, "could not parse major version")
		return
	}
	minor, err = strconv.ParseUint(versionParts[1], 10, 64)
	if err != nil {
		err = errors.Wrap(err, "could not parse minor version")
		return
	}

	patchStr := versionParts[2]
	if hyphenIndex := strings.IndexRune(patchStr, '-'); hyphenIndex != -1 {
		patchStr = patchStr[:hyphenIndex]
	}
	patch, err = strconv.ParseUint(patchStr, 10, 64)
	if err != nil {
		err = errors.Wrap(err, "could not parse patch version")
		return
	}
	return major, minor, patch, nil
}

// GetDownload returns a matching ArtifactDownload object
// given a BuildOptions object.
func (version *ArtifactVersion) GetDownload(key BuildOptions) (ArtifactDownload, error) {
	version.mutex.RLock()
	defer version.mutex.RLock()

	// TODO: this is the place to fix hanlding for the Base edition, which is not necessarily intuitive.
	if key.Edition == Base {
		if key.Target == "linux" {
			key.Target += "_" + string(key.Arch)
		}
	}

	// For OSX, the edition depends on the major/minor version.
	if key.Target == "osx" {
		major, minor, _, err := parseVersionParts(version.Version)
		if err != nil {
			return ArtifactDownload{}, errors.Wrap(err, "could not parse version")
		}
		// Before 4.1, OSX editions are "osx". However, starting in 4.1, OSX editions are "macos".
		if major > 4 || major >= 4 && minor >= 1 {
			key.Target = "macos"
		}
	}

	// we look for debug builds later in the process, but as map
	// keys, debug is always false.
	key.Debug = false

	dl, ok := version.table[key]
	if !ok {
		return ArtifactDownload{}, errors.Errorf("there is no build for %s (%s) in edition %s",
			key.Target, key.Arch, key.Edition)
	}

	return dl, nil
}

// GetBuildTypes builds, from an ArtifactsVersion object a BuildTypes
// object that reports on the available builds for this version.
func (version *ArtifactVersion) GetBuildTypes() *BuildTypes {
	out := BuildTypes{}

	seenTargets := make(map[string]struct{})
	seenEditions := make(map[MongoDBEdition]struct{})
	seenArchitectures := make(map[MongoDBArch]struct{})

	for _, dl := range version.Downloads {
		out.Version = version.Version
		if dl.Edition == "source" {
			continue
		}

		if _, ok := seenTargets[dl.Target]; !ok {
			seenTargets[dl.Target] = struct{}{}
			out.Targets = append(out.Targets, dl.Target)
		}

		if _, ok := seenEditions[dl.Edition]; !ok {
			seenEditions[dl.Edition] = struct{}{}
			out.Editions = append(out.Editions, dl.Edition)
		}

		if _, ok := seenArchitectures[dl.Arch]; !ok {
			seenArchitectures[dl.Arch] = struct{}{}
			out.Architectures = append(out.Architectures, dl.Arch)
		}
	}

	return &out
}

func (version *ArtifactVersion) String() string {
	out := []string{version.Version}

	for _, dl := range version.Downloads {
		if dl.Edition == "source" {
			continue
		}

		out = append(out, fmt.Sprintf("\t target='%s', edition='%v', arch='%v'",
			dl.Target, dl.Edition, dl.Arch))
	}

	return strings.Join(out, "\n")
}
