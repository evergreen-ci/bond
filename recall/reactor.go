package recall

import (
	"runtime"

	"github.com/mongodb/amboy/queue"
	"github.com/pkg/errors"
	"github.com/tychoish/bond"
	"github.com/tychoish/grip"
	"golang.org/x/net/context"
)

func DownloadReleases(releases []string, path string, edition bond.MongoDBEdition, arch bond.MongoDBArch, target string) error {
	feed, err := bond.NewArtifactsFeed(path)
	if err != nil {
		return errors.Wrap(err, "problem building feed")
	}

	if err := feed.Populate(); err != nil {
		return errors.Wrap(err, "problem getting feed data")
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	q := queue.NewLocalUnordered(runtime.NumCPU())
	if err := q.Start(ctx); err != nil {
		return errors.Wrap(err, "problem starting queue")
	}

	catcher := grip.NewCatcher()
	urls, errChan := feed.GetArchives(releases, edition, arch, target)
	for url := range urls {
		j, err := NewJob(url, path, false)
		if err != nil {
			catcher.Add(errors.Wrap(err,
				"problem generating task for %s", url))
			continue
		}
		if err = q.Put(j); err != nil {
			catcher.Add(errors.Wrap(err,
				"problem enquing task for %s", url))
			continue
		}
	}

	if catcher.HasErrors() {
		return catcher.Resolve()
	}

	for errs := range errChan {
		for _, err := range errs {
			catcher.Add(err)
		}
	}

	if catcher.HasErrors() {
		return catcher.Resolve()
	}

	grip.Infof("waiting for '%s' download jobs to complete", q.Stats().Total)
	q.Wait()
	grip.Info("all download tasks complete, processing errors now")

	for result := range q.Results() {
		if err := result.Error(); err != nil {
			catcher.Add(err)
		}
	}

	if catcher.HasErrors() {
		return catcher.Resolve()
	}

	return nil
}
