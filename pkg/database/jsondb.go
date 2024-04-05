package database

import (
	"encoding/json"
	"os"
	"sync"
)

// JSONDB represents the JSON database implementation.
type JSONDB struct {
	filePath string
	mutex    sync.Mutex
}

// NewJSONDB creates a new instance of JSON database.
func NewJSONDB(filePath string) *JSONDB {
	return &JSONDB{
		filePath: filePath,
	}
}

// Save saves the data to the JSON database file.
func (db *JSONDB) Save(data interface{}) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.filePath, jsonData, 0600)
	if err != nil {
		return err
	}

	return nil
}

// Load loads the data from the JSON database file.
func (db *JSONDB) Load(v interface{}) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, err := os.Stat(db.filePath); os.IsNotExist(err) {
		// If the file does not exist, return an empty slice
		return nil
	}

	data, err := os.ReadFile(db.filePath)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, v); err != nil {
		return err
	}

	return nil
}
