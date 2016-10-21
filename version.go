package bond

import (
	"sync"

	"github.com/pkg/errors"
)

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

func (version *ArtifactVersion) GetDownload(key BuildOptions) (ArtifactDownload, error) {
	version.mutex.RLock()
	defer version.mutex.RLock()

	dl, ok := version.table[key]
	if !ok {
		return ArtifactDownload{}, errors.Errorf("there is no build for %s (%s) in edition %s",
			key.Target, key.Arch, key.Edition)
	}

	return dl, nil
}

func (dl ArtifactDownload) GetPackages() []string {
	if dl.Msi != "" && len(dl.Packages) == 0 {
		return []string{dl.Msi}
	}

	return dl.Packages
}

func (dl ArtifactDownload) GetArchive() string {
	return dl.Archive.Url
}
