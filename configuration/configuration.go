package configuration

import (
	"encoding/json"
	"os"
)

type Config struct {
	HAAuthToken string `json:"ha_auth_token"`
	HAStatusURL string `json:"ha_status_url"`

	IncludeEntities []string `json:"include_entities"`
	ExcludeEntities []string `json:"exclude_entities"`
}

func Read(path string) (*Config, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var conf Config
	if err := json.Unmarshal(f, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
