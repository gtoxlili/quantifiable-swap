package config

import (
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/common"
	"gopkg.in/yaml.v3"
	"os"
)

type Symbol struct {
	Base  string `yaml:"base" json:"base"`
	Quote string `yaml:"quote" json:"quote"`
}

type Amount struct {
	Sell float64 `yaml:"sell" json:"sell"`
	Buy  float64 `yaml:"buy" json:"buy"`
}

type Provider struct {
	Name        string `yaml:"name" json:"name"`
	InjectOrder string `yaml:"inject_order,omitempty" json:"inject_order,omitempty"`
}

type Job struct {
	Id       string   `yaml:"id" json:"id"`
	Type     string   `yaml:"type" json:"type"`
	Symbol   Symbol   `yaml:"symbol" json:"symbol"`
	Bar      string   `yaml:"bar" json:"bar"`
	Amount   Amount   `yaml:"amount" json:"amount"`
	Provider Provider `yaml:"provider" json:"provider"`
}

// GetId 如果 id 不存在，则生成默认的 Id
func (j *Job) GetId() string {
	if j.Id == "" {
		return j.Provider.Name + "_" + j.Type + "_" + j.Symbol.Base + j.Symbol.Quote + "_" + j.Bar
	}
	return j.Id
}

func (j *Job) Validate(skip ...string) error {
	return common.CheckEmptyFields(j, skip...)
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
