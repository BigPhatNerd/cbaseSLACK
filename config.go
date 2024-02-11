package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Slack struct {
		BOTToken          string `yaml:"BOT_TOKEN"`
		AccessToken       string `yaml:"ACCESS_TOKEN"`
		RefreshToken      string `yaml:"REFRESH_TOKEN"`
		AppID             string `yaml:"APP_ID"`
		ClientID          string `yaml:"CLIENT_ID"`
		ClientSecret      string `yaml:"CLIENT_SECRET"`
		SigningSecret     string `yaml:"SIGNING_SECRET"`
		VerificationToken string `yaml:"VERIFICATION_TOKEN"`
	} `yaml:"slack"`
	Aws struct {
		AccessKey       string `yaml:"ACCESS_KEY"`
		SecretAccessKey string `yaml:"SECRET_ACCESS_KEY"`
	} `yaml:"aws"`
}

var configure Config


func readConfig(configPath string) error{
	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, &configure)
	if err != nil {
		return err
	}

	return nil
}