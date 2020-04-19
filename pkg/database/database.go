package database

import (
    "crypto/sha1"
    "fmt"
    "github.com/nanobox-io/golang-scribble"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "os"
    "path/filepath"
    "sync"
)

type Database struct {
    dir     string
    mutex   sync.Mutex
    mutexes map[string]*sync.Mutex
    driver  *scribble.Driver
}

type Object interface{}

func New(dir string) (database *Database, err error) {
    var driver *scribble.Driver
    driver, err = scribble.New(dir, nil)
    if err != nil {
        err = errors.Wrapv(err, "unable to create new database driver for directory", dir)
        return
    }

    database = &Database{
        dir:     dir,
        mutexes: make(map[string]*sync.Mutex),
        driver:  driver,
    }
    return
}

func (d *Database) TableExists(collection string) bool {
    dir := filepath.Join(d.dir, collection)
    _, err := os.Stat(dir)
    return !os.IsNotExist(err)
}

func (d *Database) DeleteTableIfExists(collection string) (err error) {
    mutex := d.getOrCreateMutex(collection)
    mutex.Lock()
    defer mutex.Unlock()

    if d.TableExists(collection) {
        dir := filepath.Join(d.dir, collection)
        if err = os.RemoveAll(dir); err != nil {
            return errors.Wrapv(err, "unable to delete table directory", dir)
        }
    }
    return
}

func (d *Database) write(collection, resource string, v interface{}) error {
    return d.driver.Write(collection, resource, v)
}

func (d *Database) writeIfNotExists(collection, resource string, obj interface{}) (created bool, err error) {
    mutex := d.getOrCreateMutex(collection)
    mutex.Lock()
    defer mutex.Unlock()

    var exists bool
    exists, err = d.exists(collection, resource)
    if err != nil {
        err = errors.WithMessage(err, "unable to get \"exists\" value")
        return
    }

    if !exists {
        err = d.write(collection, resource, obj)
        created = true
    }

    return
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

// getOrCreateMutex creates a new collection specific mutex any time a collection
// is being modfied to avoid unsafe operations
func (d *Database) getOrCreateMutex(collection string) *sync.Mutex {

    d.mutex.Lock()
    defer d.mutex.Unlock()

    m, ok := d.mutexes[collection]

    // if the mutex doesn't exist make it
    if !ok {
        m = &sync.Mutex{}
        d.mutexes[collection] = m
    }

    return m
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
