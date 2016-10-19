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

	table map[TargetKey]ArtifactDownload
	mutex sync.RWMutex
}

type TargetKey struct {
	Edition MongoDBEdition
	Arch    MongoDBArch
	Name    string
}

func (version *ArtifactVersion) refresh() {
	version.mutex.Lock()
	defer version.mutex.Unlock()

	version.table = make(map[TargetKey]ArtifactDownload)

	for _, dl := range version.Downloads {
		key := TargetKey{
			Edition: dl.Edition,
			Name:    dl.Target,
			Arch:    dl.Arch,
		}

		version.table[key] = dl
	}
}

func (version *ArtifactVersion) GetDownload(edition MongoDBEdition, arch MongoDBArch, target string) (ArtifactDownload, error) {
	version.mutex.RLock()
	defer version.mutex.RLock()

	key := TargetKey{
		Edition: edition,
		Name:    target,
		Arch:    arch,
	}

	dl, ok := version.table[key]
	if !ok {
		return ArtifactDownload{}, errors.Errorf("there is no build for %s (%s) in edition %s",
			target, arch, edition)
	}

	return dl, nil
}

func (dl ArtifactDownload) getPackages() []string {
	if dl.Msi != "" && len(dl.Packages) == 0 {
		return []string{dl.Msi}
	}

	return dl.Packages
}

//
// Base (generic) Accessors
//

func (version *ArtifactVersion) GetBaseArchive(target string, arch MongoDBArch) (string, error) {
	build, err := version.GetDownload(Base, arch, target)
	if err != nil {
		return "", err
	}

	return build.Archive.Url, nil
}

func (version *ArtifactVersion) GetBasePackages(target string, arch MongoDBArch) ([]string, error) {
	build, err := version.GetDownload(Base, arch, target)
	if err != nil {
		return []string{}, err
	}

	return build.getPackages(), nil
}

//
// Enterprise Accessors
//

func (version *ArtifactVersion) GetEnterpriseArchive(target string, arch MongoDBArch) (string, error) {
	build, err := version.GetDownload(Enterprise, arch, target)
	if err != nil {
		return "", err
	}

	return build.Archive.Url, nil
}

func (version *ArtifactVersion) GetEnterprisePackages(target string, arch MongoDBArch) ([]string, error) {
	build, err := version.GetDownload(Enterprise, arch, target)
	if err != nil {
		return []string{}, err
	}

	return build.getPackages(), nil
}

//
// Community Targeted Accessors
//

func (version *ArtifactVersion) GetCommunityArchive(target string, arch MongoDBArch) (string, error) {
	build, err := version.GetDownload(CommunityTargeted, arch, target)
	if err != nil {
		return "", err
	}

	return build.Archive.Url, nil
}

func (version *ArtifactVersion) GetCommunityPackages(target string, arch MongoDBArch) ([]string, error) {
	build, err := version.GetDownload(CommunityTargeted, arch, target)
	if err != nil {
		return []string{}, err
	}

	return build.getPackages(), nil
}
