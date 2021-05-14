package datastore

import (
	"io"

	"gopkg.in/yaml.v3"
)

type YAMLDataStore struct{}

func (Y YAMLDataStore) Store(bind map[string]interface{}, w io.Writer) error {
	if b, err := yaml.Marshal(bind); err != nil {
		return err
	} else {
		if _, err := w.Write(b); err != nil {
			return err
		}
	}
	return nil
}

func (Y YAMLDataStore) Open(r io.Reader, bind map[string]interface{}) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, &bind)
}
