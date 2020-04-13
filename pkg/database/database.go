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

func New(dir string) (database *Database, err error) {
    driver, err := scribble.New(dir, nil)
    if err != nil {
        return
    }

    database = &Database{
        dir:    dir,
        driver: driver,
    }
    return
}

func (d *Database) TableExists(table string) bool {
    dir := filepath.Join(d.dir, table)
    _, err := os.Stat(dir)
    return !os.IsNotExist(err)
}

func (d *Database) DeleteTableIfExists(table string) (err error) {
    if d.TableExists(table) {
        dir := filepath.Join(d.dir, table)
        if err = os.RemoveAll(dir); err != nil {
            return errors.Wrapv(err, "unable to delete table directory", dir)
        }
    }
    return
}

func (d *Database) write(collection, resource string, v interface{}) error {
    return d.driver.Write(collection, resource, v)
}

func (d *Database) read(collection, resource string, v interface{}) (err error) {
    if !d.TableExists(collection) {
        return
    }
    return d.driver.Read(collection, resource, v)
}

func (d *Database) exists(collection, resource string) (result bool, err error) {
    if !d.TableExists(collection) {
        return
    }

    if readErr := d.driver.Read(collection, resource, nil); readErr != nil {
        if os.IsNotExist(readErr) {
            return
        }

        err = readErr
        return
    }

    result = true

    return
}

func (d *Database) delete(collection, resource string) error {
    return d.driver.Delete(collection, resource)
}

func (d *Database) readAll(collection string) (rows []string, err error) {
    if !d.TableExists(collection) {
        return
    }
    return d.driver.ReadAll(collection)
}

func CreateHashID(firstInput interface{}, otherInputs ...interface{}) (result string) {
    str := fmt.Sprintf("%v", firstInput)
    for _, otherInput := range otherInputs {
        str += fmt.Sprintf("-%v", otherInput)
    }

    h := sha1.New()
    h.Write([]byte(str))
    bs := h.Sum(nil)

    return fmt.Sprintf("%x", bs)
}
