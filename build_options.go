package bond

import (
	"encoding/json"

	"github.com/mongodb/grip"
)

// BuildOptions is a common method to describe a build variant.
type BuildOptions struct {
	Target  string         `json:"target"`
	Arch    MongoDBArch    `json:"arch"`
	Edition MongoDBEdition `json:"edition"`
	Debug   bool           `json:"debug"`
}

func (o BuildOptions) String() string {
	out, err := json.Marshal(o)
	if err != nil {
		return "{}"
	}
	return string(out)
}

// GetBuildInfo given a version string, generates a BuildInfo object
// from a BuildOptions object.
func (o BuildOptions) GetBuildInfo(version string) BuildInfo {
	return BuildInfo{
		Version: version,
		Options: o,
	}
}

// Validate checks a BuildOption structure and ensures that there are
// no errors.
func (o BuildOptions) Validate() error {
	catcher := grip.NewBasicCatcher()

	catcher.NewWhen(o.Target == "", "missing target")
	catcher.NewWhen(o.Arch == "", "missing arch")
	catcher.NewWhen(o.Edition == "", "missing edition")

	return catcher.Resolve()
}
