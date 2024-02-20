package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Slack struct {
		BOTToken          string `yaml:"BOT_TOKEN"`
		AppAccessToken       string `yaml:"APP_CONFIG_ACCESS_TOKEN"`
		AppRefreshToken      string `yaml:"APP_REFRESH_TOKEN"`
		AppID             string `yaml:"APP_ID"`
		ClientID          string `yaml:"CLIENT_ID"`
		ClientSecret      string `yaml:"CLIENT_SECRET"`
		SigningSecret     string `yaml:"SIGNING_SECRET"`
		VerificationToken string `yaml:"VERIFICATION_TOKEN"`
		TeamID            string `yaml:"TEAM_ID"`
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