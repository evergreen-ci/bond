package bond

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

type ReleaseInfo struct {
	mu   sync.RWMutex
	data map[string]string
}

func CollectReleaseInfo() (*ReleaseInfo, error) {
	files, err := filepath.Glob("/etc/*release")
	if err != nil {
		return nil, errors.Wrap(err, "problem finding release file")
	}

	if len(files) == 0 {
		return nil, errors.New("found no matching release file")
	}

	info := &ReleaseInfo{data: map[string]string{}}
	catcher := grip.NewBasicCatcher()
	for _, fn := range files {
		func() {
			file, err := os.Open(fn)
			if err != nil {
				catcher.Add(err)
				return
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				parts := strings.SplitN(line, "=", 2)
				if len(parts) != 2 {
					catcher.Add(errors.Errorf("found invalid line '%s'", line))
					continue
				}

				info.data[strings.ToLower(parts[0])] = strings.Trim(parts[1], "\"'")
			}
			catcher.Add(scanner.Err())
		}()
	}

	if catcher.HasErrors() {
		return nil, catcher.Resolve()
	}

	return info, nil
}
