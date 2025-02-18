package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Job struct {
	Id     string `yaml:"id"`
	Type   string `yaml:"type"`
	Symbol struct {
		Base  string `yaml:"base"`
		Quote string `yaml:"quote"`
	} `yaml:"symbol"`
	Bar    string `yaml:"bar"`
	Amount struct {
		Sell string `yaml:"sell"`
		Buy  string `yaml:"buy"`
	} `yaml:"amount"`
	Provider struct {
		Name        string `yaml:"name"`
		InjectOrder string `yaml:"inject_order,omitempty"`
	} `yaml:"provider"`
}

func ParseConfig(path string) ([]Job, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var jobs []Job
	if err := yaml.NewDecoder(file).Decode(&jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}
