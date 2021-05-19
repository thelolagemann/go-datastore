// Package datastore provides a relatively simple flat data
// store using key, value pairs, which can then be saved to
// an implementation of Storer.
package datastore

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	// JSONStore store data in JSON format
	JSONStore StoreType = iota
	// YAMLStore store data in YAML format
	YAMLStore
)

// StoreType the type of store
type StoreType int


type NoRecordError struct {
	key string
	store string
}

func (n NoRecordError) Error() string {
	return fmt.Sprintf("key %v doesn't exist in the %v store", n.key, n.store)
}

// Storer is the interface that wraps an implementation of a data
// store. Implementations must be capable of both retrieving and
// storing raw data, via the Open and Store methods respectively.
type Storer interface {
	Open(r io.Reader, out map[string]interface{}) error
	Store(in map[string]interface{}, w io.Writer) error
}

// Store is the base object that values are read and written
// from.
type Store struct {
	*Config
	Storer

	records map[string]interface{}
	f *os.File
}

// Delete deletes the record with key from the data store. If
// no record can be found, an ErrNoKey will be returned.
func (d *Store) Delete(key string) error {
	if _, ok := d.records[key]; ok {
		delete(d.records, key)
		return nil
	}
	return NoRecordError{key, d.StoreName}
}

// Read reads the record from the data store with key, and marshals
// the value into bind. If no record can be found, an ErrNoKey
// will be returned.
func (d *Store) Read(key string, bind interface{}) error {
	if v, ok := d.records[key]; ok {
		return marshalRecords(v, &bind)
	}

	return NoRecordError{key, d.StoreName}
}

// ReadAll reads all of the values from d.records, and marshals the
// values into bind.
func (d *Store) ReadAll(bind interface{}) error {
	return marshalRecords(d.records, bind)
}

// Write writes the value of bind into the data store with key.
func (d *Store) Write(key string, bind interface{}) error {
	d.records[key] = bind
	if d.SaveOnWrite {
		return d.save()
	}
	return nil
}

// Close flushes the contents of the data store to the underlying
// file, and then closes any remaining file handles.
func (d *Store) Close() error {
	if err := d.save(); err != nil {
		return err
	}
	return d.f.Close()
}

func marshalRecords(in interface{}, out interface{}) error {
	// use json to marshal and unmarshal
	// TODO custom marshaller
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func (d *Store) save() error {
	// write to temp file in case error
	f, err := os.OpenFile(filepath.Join(os.TempDir(), d.f.Name()), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	if err := d.Store(d.records, f); err != nil {
		return err
	}

	// copy temp file
	if err := d.f.Truncate(0); err != nil {
		return err
	}
	if _, err := d.f.Seek(0, 0); err != nil {
		return err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	if _, err := io.Copy(d.f, f); err != nil {
		return err
	}

	return nil
}

type Config struct {
	// SaveOnWrite if true every write call will flush the store to stable storage
	 SaveOnWrite bool
	 StoreType

	 // StoreName is the name assigned to the store. If no name is provided,
	 // then the name will be derived from the file name.
	 StoreName string

	 // Log generic logging interface
	 Log Logger
}

// New creates or opens an existing Store of type d, at the provided
// path. If storeOnWrite is true, every Store.Write() call will flush the
// contents of the data store to disk.
func New(path string, config *Config) (*Store, error) {
	// open the store
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	var dS = &Store{
		Config: config,
		f:            f,
		records:      map[string]interface{}{},
	}

	switch config.StoreType {
	case JSONStore:
		dS.Storer = JSONDataStore{}
	case YAMLStore:
		dS.Storer = YAMLDataStore{}

	default:
		return nil, fmt.Errorf("unsupported store type: %v", config.StoreType)
	}

	// read records into store
	if err := dS.Open(f, dS.records); err != nil {
		return nil, err
	}

	// setup logging
	if config.Log == nil {
		config.Log = &noLogger{}
	}

	// configure name
	if config.StoreName == "" {
		config.StoreName = dS.f.Name()
	}

	return dS, nil
}
