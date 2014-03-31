package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Config struct {
	AllowedEntries []Entry `json:allowedEntries`
}

const (
	TypeResize = "resize"
	TypeCut    = "cut"
)

type Entry struct {
	Name   string `json:name`
	Width  int64  `json:width`
	Height int64  `json:height`
	Type   string `json:type`
}

// CreateConfigFromFile returns an Config object from a given file.
func CreateConfigFromFile(file string) (*Config, error) {
	result := Config{}

	config, err := ioutil.ReadFile(file)

	if err != nil {
		return &result, err
	}

	err = json.Unmarshal(config, &result)

	err = result.validateConfig()

	return &result, err
}

// validateConfig validates the configuration and fills elements with default types.
func (config *Config) validateConfig() error {
	for index, element := range config.AllowedEntries {
		if element.Width <= 0 && element.Height <= 0 {
			return fmt.Errorf("The width and height of the configuration element with name \"%s\" are invalid.", element.Name)
		}

		if element.Type == "" {
			config.AllowedEntries[index].Type = TypeResize
			continue
		}

		if element.Type != TypeResize && element.Type != TypeCut {
			return fmt.Errorf("Type must be either %s or %s at element \"%s\"", TypeCut, TypeResize, element.Name)
		}
	}

	return nil
}

// Returns an entry the the name.
func (config *Config) GetEntryByName(name string) (*Entry, error) {
	for _, element := range config.AllowedEntries {
		if element.Name == name {
			return &element, nil
		}
	}

	return nil, fmt.Errorf("No Entry found in configuration for given name %s", name)
}
