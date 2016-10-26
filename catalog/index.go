package catalog

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/tychoish/grip"
)

func Show(path string) error {
	var err error
	if _, err = os.Stat(path); err != nil {
		return errors.Errorf("path %s does not exist", path)
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return errors.Wrap(err, "problem resolving absolute path")
	}

	contents, err := ioutil.ReadDir(path)
	if err != nil {
		return errors.Wrapf(err, "problem fetching contents of %s", path)
	}

	if len(contents) == 0 {
		return errors.Errorf("path %s is empty", path)
	}

	catcher := grip.NewCatcher()
	// map of versions to abspaths
	cache := map[string]string{}
	for _, info := range contents {
		if !strings.HasPrefix(info.Name(), "mongodb-") {
			grip.Debugf("skipping file %s, which is not a mongodb artifact", info.Name())
			continue
		}

		if !info.IsDir() {
			grip.Debugf("'%s' is a file, likely an archive", info.Name())
			continue
		}

		fqfn := filepath.Join(path, info.Name())
		version := getVersion(info.Name())

		grip.Infoln("got version", version, "from", info.Name())

		cache[version] = fqfn
		catcher.Add(validatePackageContents(fqfn))
	}

	for k, v := range cache {
		fmt.Println(k, "->", v)
	}

	grip.AlertWhen(catcher.HasErrors(), catcher.Resolve())

	return nil
}

func validatePackageContents(path string) error {
	path = filepath.Join(path, "bin")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.Errorf("directory %s does not exist, not a valid package", path)
	}

	contents, err := ioutil.ReadDir(path)
	if err != nil {
		return errors.Wrapf(err, "problem finding contents of %s")
	}

	pkg := make(map[string]struct{})
	for _, info := range contents {
		pkg[info.Name()] = struct{}{}
	}

	errs := []string{}

	for _, bin := range []string{"mongod", "mongos"} {
		if runtime.GOOS == "windows" {
			bin += ".exe"
		}

		if _, ok := pkg[bin]; !ok {
			errs = append(errs, fmt.Sprintf("binary %s is missing from %s", bin, path))
		}
	}

	if len(errs) >= 1 {
		return errors.New(strings.Join(errs, ";"))
	}

	return nil
}

func getVersion(name string) string {
	fnParts := strings.Split(name, "-")
	// this won't get nightlies correctly
	_, err := strconv.Atoi(fnParts[len(fnParts)-2])
	if strings.Contains(name, "rc") {
		if err == nil {
			// nightly of an rc
			return strings.Join(fnParts[len(fnParts)-4:], "-")
		}
		// any other rc.
		return strings.Join(fnParts[len(fnParts)-2:], "-")
	} else if strings.Contains(name, "v2.4") {
		// v2.4-latest has different format.
		return strings.Join(fnParts[len(fnParts)-4:], "-")[1:]
	} else if err == nil {
		// most new-style nightlies.
		return strings.Join(fnParts[len(fnParts)-3:], "-")
	}

	// on GA releases, which is by far the most
	// common, the version is the last element.
	return fnParts[len(fnParts)-1]
}
