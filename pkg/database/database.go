package database

import (
	"crypto/sha1"
	"fmt"
	"github.com/nanobox-io/golang-scribble"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	"os"
	"path/filepath"
)

type Database struct {
	dir    string
	driver *scribble.Driver
}

type Object interface{}

func New(dir string) (*Database, error) {
	driver, err := scribble.New(dir, nil)
	if err != nil {
		return nil, errors.Wrapv(err, "unable to create scribble driver", dir)
	}

	return &Database{
		dir:    dir,
		driver: driver,
	}, nil
}

func (d *Database) TableExists(table string) bool {
	dir := filepath.Join(d.dir, table)
	_, err := os.Stat(dir)
	return !os.IsNotExist(err)
}

// Wrapped
func (d *Database) Write(collection, resource string, v interface{}) error {
	return d.driver.Write(collection, resource, v)
}

// Wrapped
func (d *Database) Read(collection, resource string, v interface{}) error {
	return d.driver.Read(collection, resource, v)
}

// Wrapped
func (d *Database) ReadAll(collection string) ([]string, error) {
	return d.driver.ReadAll(collection)
}

func CreateHashID(input string) string {
	h := sha1.New()
	h.Write([]byte(input))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
