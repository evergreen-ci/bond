package recall

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/dependency"
	"github.com/mongodb/amboy/registry"
	"github.com/pkg/errors"
	"github.com/tychoish/bond"
	"github.com/tychoish/grip"
)

type DownloadFileJob struct {
	URL            string `bson:"url" json:"url" yaml:"url"`
	Directory      string `bson:"dir" json:"dir" yaml:"dir"`
	FileName       string `bson:"file" json:"file" yaml:"file"`
	*amboy.JobBase `bson:"metadata" json:"metadata" yaml:"metadata"`
}

func init() {
	registry.AddJobType("bond-recall-download-file", func() amboy.Job {
		return newDownloadJob()
	})
}

func newDownloadJob() *DownloadFileJob {
	return &DownloadFileJob{
		JobBase: &amboy.JobBase{
			JobType: amboy.JobType{
				Name:    "bond-recall-download-file",
				Version: 0,
				Format:  amboy.JSON,
			},
		},
	}
}

func NewJob(url, path string, force bool) (*DownloadFileJob, error) {
	j := newDownloadJob()
	if err := j.setURL(url); err != nil {
		return nil, errors.Wrap(err, "problem constructing Job object (url)")
	}

	if err := j.setDirectory(url); err != nil {
		return nil, errors.Wrap(err, "problem constructing Job object (directory)")
	}

	if force {
		j.SetDependency(dependency.NewAlways())
	} else {
		j.SetDependency(dependency.NewCreatesFile(j.getFileName()))
	}

	return j, nil
}

func (j *DownloadFileJob) Run() {
	defer j.MarkComplete()

	fn := j.getFileName()

	// in theory the queue should do this next check, but most do not
	if state := j.Dependency().State(); state == dependency.Passed {
		grip.Noticef("file %s is already downloaded", fn)
		return
	}

	if err := bond.DownloadFile(j.URL, fn); err != nil {
		err = errors.Wrapf(err, "problem downloading file %s", fn)
		j.AddError(err)
		grip.CatchError(err)
		return
	}

	grip.Noticef("downloaded %s file", fn)
}

//
// Internal Methods
//

func (j *DownloadFileJob) getFileName() string {
	return filepath.Join(j.Directory, j.FileName)
}

func (j *DownloadFileJob) setDirectory(path string) error {
	if stat, err := os.Stat(path); !os.IsNotExist(err) && !stat.IsDir() {
		// if the path exists and isn't a directory, then we
		// won't be able to download into it:
		return errors.Errorf("%s is not a directory, cannot download files into it",
			path)
	}

	j.Directory = path
	return nil
}

func (j *DownloadFileJob) setURL(url string) error {
	if strings.HasPrefix(url, "http") {
		j.URL = url
		j.FileName = filepath.Base(url)
	}

	return errors.Errorf("%s is not a valid url", url)
}
