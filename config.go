package main

import (
	"flag"
	"gopkg.in/yaml.v3"
	"os"
)

type Configuration struct {
	Cognito struct {
		Region       string `yaml:"region"`
		ClientID     string `yaml:"clientID"`
		ClientSecret string `yaml:"clientSecret"`
	} `yaml:"cognito"`
	Service struct {
		MolecularURL string `yaml:"molecularURL"`
		DexURL       string `yaml:"dexURL"`
	} `yaml:"service"`
	Frontend struct {
		Directory string `yaml:"directory"`
	} `yaml:"frontend"`
}

func (c *Configuration) Get() error {
	var configFile string
	flag.StringVar(&configFile, "config", "./.conf/config.dev.yaml", "path to config file")
	flag.Parse()

	configData, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(configData, c); err != nil {
		return err
	}

	return nil
}
