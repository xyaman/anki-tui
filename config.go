package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

const APPNAME = "ankimorph"
const FILENAME = "config.yaml"

type Config struct {
	Query string `yaml:"query"`

	// All fields can be separaed by comma, and the program will use the first one that is not nil
	MorphFieldName    string `yaml:"morphFieldName"`
	SentenceFieldName string `yaml:"sentenceFieldName"`
	ImageFieldName    string `yaml:"imageFieldName"`
	AudioFieldName    string `yaml:"audioFieldName"`
	KnownTag          string `yaml:"knownTag"`

	MinningImageFieldName string `yaml:"minningImageFieldName"`
	MinningAudioFieldName string `yaml:"minningAudioFieldName"`

	PlayAudioAutomatically bool `yaml:"playAudioAutomatically"`
}

func loadConfig() (*Config, error) {

	var configPath string
	var config *Config

	// Get os name
	switch runtime.GOOS {
	case "windows":
		configPath = os.Getenv("APPDATA")
	case "linux":
		if os.Getenv("XDG_CONFIG_HOME") != "" {
			configPath = os.Getenv("XDG_CONFIG_HOME")
		} else {
			configPath = filepath.Join(os.Getenv("HOME") + "/.config")
		}
	default:
		return nil, fmt.Errorf("unsupported os: %s", runtime.GOOS)
	}

	// If file exists load it
	if _, err := os.Stat(filepath.Join(configPath, APPNAME, FILENAME)); err == nil {
		// read file
		data, err := os.ReadFile(filepath.Join(configPath, APPNAME, FILENAME))
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return nil, err
		}
	} else {
		// Create file
		err := os.MkdirAll(filepath.Join(configPath, APPNAME), 0755)
		if err != nil {
			return nil, err
		}

		config = &Config{
			Query:             "deck:ankimorph tag:1T -tag:MT",
			MorphFieldName:    "am-unknowns",
			SentenceFieldName: "Expression",
			ImageFieldName:    "Screenshot",
			AudioFieldName:    "Audio_Sentence",
			KnownTag:          "am-known-manually",

			MinningImageFieldName: "Picture",
			MinningAudioFieldName: "SentenceAudio",

			PlayAudioAutomatically: false,
		}

		data, err := yaml.Marshal(config)
		if err != nil {
			return nil, err
		}

		err = os.WriteFile(filepath.Join(configPath, APPNAME, FILENAME), data, 0755)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

func (c *Config) save() error {
	var configPath string

	// Get os name
	switch runtime.GOOS {
	case "windows":
		configPath = os.Getenv("APPDATA")
	case "linux":
		if os.Getenv("XDG_CONFIG_HOME") != "" {
			configPath = os.Getenv("XDG_CONFIG_HOME")
		} else {
			configPath = filepath.Join(os.Getenv("HOME") + "/.config")
		}
	default:
		return fmt.Errorf("unsupported os: %s", runtime.GOOS)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(configPath, APPNAME, FILENAME), data, 0755)
	if err != nil {
		return err
	}

	return nil
}
