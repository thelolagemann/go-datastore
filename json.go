package datastore

import (
	"encoding/json"
	"io"
)

type JSONDataStore struct{}

func (j JSONDataStore) Open(r io.Reader, bind map[string]interface{}) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	// new file
	if len(b) == 0 {
		b = []byte("{}")
	}

	return json.Unmarshal(b, &bind)
}

func (j JSONDataStore) Store(bind map[string]interface{}, w io.Writer) error {
	// indent for easy reading
	// TODO make configurable
	if b, err := json.MarshalIndent(bind, "", "\t"); err != nil {
		return err
	} else {
		if _, err := w.Write(b); err != nil {
			return err
		}
	}

	return nil
}