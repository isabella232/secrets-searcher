package build

import (
	"os"
	"path/filepath"

	"github.com/pantheon-systems/search-secrets/pkg/errors"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
)

func AddSecretFilesFromDirectory(path string, secretIDs manip.Set) (err error) {
	if err = filepath.Walk(path, addSecretFileFunc(secretIDs)); err != nil {
		err = errors.Wrap(err, "unable to add secret IDs from whitelist directory")
	}
	return
}

func addSecretFileFunc(secretIDs manip.Set) func(filePath string, info os.FileInfo, err error) error {
	return func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		matches := SecretFileMatch.FindStringSubmatch(info.Name())
		if len(matches) == 0 {
			return nil
		}
		secretIDs.Add(matches[1])
		return nil
	}
}
