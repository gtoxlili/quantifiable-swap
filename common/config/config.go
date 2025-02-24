package config

import (
	"encoding/json"
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/common"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

type Symbol struct {
	Base  string `yaml:"base" json:"base"`
	Quote string `yaml:"quote" json:"quote"`
}

type Amount struct {
	Sell float64 `yaml:"sell,omitempty" json:"sell,omitempty"`
	Buy  float64 `yaml:"buy,omitempty" json:"buy,omitempty"`
}

type Provider struct {
	Data    string `yaml:"data" json:"data"`
	Trading string `yaml:"trading,omitempty" json:"trading,omitempty"`
}

type Subscriber struct {
	ID            int64 `yaml:"id" json:"id"`
	ImportantOnly bool  `yaml:"important_only" json:"important_only"`
}

type Job struct {
	Type        string       `yaml:"type" json:"type"`
	Symbol      Symbol       `yaml:"symbol" json:"symbol"`
	Bar         string       `yaml:"bar" json:"bar"`
	Amount      Amount       `yaml:"amount,omitempty" json:"amount,omitempty"`
	Provider    Provider     `yaml:"provider" json:"provider"`
	Subscribers []Subscriber `yaml:"subscribers,omitempty" json:"subscribers,omitempty"`
}

func (j *Job) String() string {
	provider := j.Provider.Data
	if j.Provider.Trading != "" && j.Provider.Trading != provider {
		provider += "(" + j.Provider.Trading + ")"
	}
	return strings.ToUpper(provider + "·" + j.Type + "·" + j.Symbol.Base + "/" + j.Symbol.Quote + "·" + j.Bar)
}

func (j *Job) Validate(skip ...string) error {
	return common.CheckEmptyFields(j, skip...)
}

func (j *Job) Format() []byte {
	format, _ := json.MarshalIndent(j, "", "  ")
	return format
}

func ParseConfig(path string) ([]Job, error) {
	if path == "" {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var jobs []Job
	if err := yaml.Unmarshal(data, &jobs); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return jobs, nil
}

func SaveConfig(path string, jobs []Job) error {
	// Create file with read/write permissions for owner only
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	encoder := yaml.NewEncoder(file)
	defer func() {
		_ = encoder.Close()
		_ = file.Close()
	}()

	if err := encoder.Encode(jobs); err != nil {
		return err
	}
	return file.Sync()
}
