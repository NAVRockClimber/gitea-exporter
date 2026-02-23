package prom

import (
	"os"

	"go.yaml.in/yaml/v2"
)

// Targets is map of probing targets
type Targets map[string]Server

// Server struct describe a gitea server
type Server struct {
	URL          string   `yaml:"url"`
	Token        string   `yaml:"token"`
	TokenEnvName string   `yaml:"tokenEnvName"`
	ExcludeOrgs  []string `yaml:"excludeOrgs"`
}

func readConfig(configFile string) Targets {
	file, err := os.Open(configFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var targets Targets
	d := yaml.NewDecoder(file)
	if err = d.Decode(&targets); err != nil {
		panic(err)
	}

	for key, v := range targets {
		if v.TokenEnvName != "" {
			envToken := os.Getenv(v.TokenEnvName)
			if envToken != "" {
				v.Token = envToken
				targets[key] = v
			}
		}
	}
	return targets
}
