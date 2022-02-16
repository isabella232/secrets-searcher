package database

// Took a lot of this from [Scribble](https://github.com/nanobox-io/golang-scribble).
// License can be found there.

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/pantheon-systems/secrets-searcher/pkg/errors"
	"github.com/pantheon-systems/secrets-searcher/pkg/logg"
)

type Database struct {
	dir     string
	mutex   *sync.Mutex
	mutexes map[string]*sync.Mutex
	log     logg.Logg
}

func New(dir string, log logg.Logg) (result *Database, err error) {
	dir = filepath.Clean(dir)

	result = &Database{
		dir:     dir,
		mutex:   &sync.Mutex{},
		mutexes: make(map[string]*sync.Mutex),
		log:     log,
	}

	return
}

func (d *Database) PrepareFilesystemForWriting() (err error) {
	if _, statErr := os.Stat(d.dir); !os.IsNotExist(statErr) {
		d.log.Debug("Using existing database directory: ", d.dir)
		return
	}

	d.log.Debugf("Creating scribble database at '%s'", d.dir)
	err = os.MkdirAll(d.dir, 0755)

	return
}

//
// Check

func (d *Database) tableExists(collection string) bool {
	d.lockTable(collection)
	defer d.unlockTable(collection)

	return d.tableExistsUnsafe(collection)
}

func (d *Database) tableExistsUnsafe(collection string) bool {
	dir := filepath.Join(d.dir, collection)
	_, err := os.Stat(dir)

	return !os.IsNotExist(err)
}

func (d *Database) exists(collection, resource string) (result bool) {
	d.lockTable(collection)
	defer d.unlockTable(collection)

	return d.existsUnsafe(collection, resource)
}

func (d *Database) existsUnsafe(collection, resource string) (result bool) {
	if !d.tableExistsUnsafe(collection) {
		return
	}

	file := filepath.Join(d.dir, collection, resource+".json")
	_, err := os.Stat(file)

	result = !os.IsNotExist(err)

	return
}

//
// Read

func (d *Database) read(collection, resource string, v interface{}) (err error) {
	if collection == "" {
		panic("no collection passed")
	}
	if resource == "" {
		panic("no resource passed")
	}

	d.lockTable(collection)
	defer d.unlockTable(collection)

	if !d.tableExistsUnsafe(collection) {
		return errors.Errorf("%s collection does not exist", collection)
	}
	if !d.existsUnsafe(collection, resource) {
		return errors.Errorf("%s record %s does not exist", collection, resource)
	}

	file := filepath.Join(d.dir, collection, resource+".json")

	var b []byte
	b, err = ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, &v)
}

func (d *Database) readAll(collection string) (result []string, err error) {
	d.lockTable(collection)
	defer d.unlockTable(collection)

	return d.readAllUnsafe(collection)
}

func (d *Database) readAllUnsafe(collection string) (result []string, err error) {
	if collection == "" {
		panic("no collection passed")
	}

	if !d.tableExistsUnsafe(collection) {
		return
	}

	dir := filepath.Join(d.dir, collection)
	files, _ := ioutil.ReadDir(dir)
	result = make([]string, len(files))
	for i, file := range files {
		var b []byte
		name := file.Name()
		path := filepath.Join(dir, name)
		b, err = ioutil.ReadFile(path)
		if err != nil {
			return
		}

		result[i] = string(b)
	}

	return
}

//
// Delete

func (d *Database) deleteTable(collection string) (err error) {
	d.lockTable(collection)
	defer d.unlockTable(collection)

	return d.deleteTableUnsafe(collection)
}

func (d *Database) deleteTables(collections []string) (err error) {
	d.lockTables(collections)
	defer d.unlockTables(collections)

	for _, collection := range collections {
		if err = d.deleteTableUnsafe(collection); err != nil {
			return errors.WithMessagev(err, "unable to delete table", collection)
		}
	}

	return
}

func (d *Database) deleteTableUnsafe(collection string) (err error) {
	if !d.tableExistsUnsafe(collection) {
		return
	}

	dir := filepath.Join(d.dir, collection)
	if err = os.RemoveAll(dir); err != nil {
		return errors.Wrapv(err, "unable to delete table directory", dir)
	}

	return
}

func (d *Database) delete(collection, resource string) (err error) {
	d.lockTable(collection)
	defer d.unlockTable(collection)

	if !d.existsUnsafe(collection, resource) {
		return
	}

	filePath := filepath.Join(d.dir, collection, resource+".json")
	err = os.RemoveAll(filePath)

	return
}

// Write

func (d *Database) writeIfNotExists(collection, resource string, obj interface{}) (created bool, err error) {
	d.lockTable(collection)
	defer d.unlockTable(collection)

	if d.existsUnsafe(collection, resource) {
		return
	}

	err = d.writeUnsafe(collection, resource, obj)
	created = true

	return
}

func (d *Database) write(collection, resource string, v interface{}) (err error) {
	d.lockTable(collection)
	defer d.unlockTable(collection)

	return d.writeUnsafe(collection, resource, v)
}

func (d *Database) writeUnsafe(collection, resource string, v interface{}) (err error) {
	if collection == "" {
		panic("Missing collection - no place to save record!")
	}
	if resource == "" {
		panic("Missing resource - unable to save record (no name)!")
	}

	dir := filepath.Join(d.dir, collection)
	fnlPath := filepath.Join(dir, resource+".json")
	tmpPath := fnlPath + ".tmp"

	if err = os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var b []byte
	b, err = json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(tmpPath, b, 0644); err != nil {
		return err
	}

	d.log.Tracef("creating %s", filepath.Join(collection, resource+".json"))
	return os.Rename(tmpPath, fnlPath)
}

//
// Locking

func (d *Database) lockTables(collections []string) {
	for _, collection := range collections {
		d.lockTable(collection)
	}
}

func (d *Database) unlockTables(collections []string) {
	for _, collection := range collections {
		d.unlockTable(collection)
	}
}

func (d *Database) lockTable(collection string) {
	d.log.Tracef("locking %s", collection)
	d.getOrCreateMutex(collection).Lock()
}

func (d *Database) unlockTable(collection string) {
	d.log.Tracef("unlocking %s", collection)
	d.getOrCreateMutex(collection).Unlock()
}

func (d *Database) getOrCreateMutex(collection string) *sync.Mutex {
	d.log.Tracef("locking main mutex %s", collection)
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
