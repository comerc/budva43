package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ghodss/yaml"
	"github.com/radovskyb/watcher"
)

var path = filepath.Join(".", filename)

func Load() (*Config, error) {
	var (
		configData Config
		err        error
		file       *os.File
		yamlData   []byte
		jsonData   []byte
	)

	file, err = os.Open(path)
	if err != nil {
		log.Printf("Failed to open file %s: %s", path, err)
		return nil, err
	}
	defer file.Close()

	yamlData, err = ioutil.ReadAll(file)
	if err != nil {
		log.Printf("Failed to read file %s: %s", path, err)
		return nil, err
	}

	jsonData, err = yaml.YAMLToJSON(yamlData)
	if err != nil {
		log.Printf("Failed to convert file %s with YAMLToJSON: %s", path, err)
		return nil, err
	}

	err = json.Unmarshal(jsonData, &configData)
	if err != nil {
		log.Printf("Failed to unmarshal file %s: %s", path, err)
		return nil, err
	}

	return &configData, err
}

func Watch(reload func()) {
	w := watcher.New()

	// SetMaxEvents to 1 to allow at most 1 event's to be received
	// on the Event channel per watching cycle.
	//
	// If SetMaxEvents is not set, the default is to send all events.
	w.SetMaxEvents(1)

	// Only notify write events.
	w.FilterOps(watcher.Write)

	go func() {
		for {
			select {
			case event := <-w.Event:
				log.Print(event) // Print the event's info.
				_ = event
				reload()
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Watch this path for changes.
	if err := w.Add(path); err != nil {
		log.Fatalln(err)
	}

	reload()

	// Start the watching process - it'll check for changes every 1000ms.
	if err := w.Start(1000 * time.Millisecond); err != nil {
		log.Fatalln(err)
	}
}
