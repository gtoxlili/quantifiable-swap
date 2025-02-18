package config

import (
	"github.com/gtoxlili/quantifiable-swap/common"
	"gopkg.in/yaml.v3"
	"os"
)

type Job struct {
	Id     string `yaml:"id" json:"id"`
	Type   string `yaml:"type" json:"type"`
	Symbol struct {
		Base  string `yaml:"base" json:"base"`
		Quote string `yaml:"quote" json:"quote"`
	} `yaml:"symbol" json:"symbol"`
	Bar    string `yaml:"bar" json:"bar"`
	Amount struct {
		Sell float64 `yaml:"sell" json:"sell"`
		Buy  float64 `yaml:"buy" json:"buy"`
	} `yaml:"amount" json:"amount"`
	Provider struct {
		Name        string `yaml:"name" json:"name"`
		InjectOrder string `yaml:"inject_order,omitempty" json:"inject_order,omitempty"`
	} `yaml:"provider" json:"provider"`
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
