package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config is the root of the application's configuration, loaded from
// a YAML file and environment variables.
type Config struct {
	ListenAddress  string            `yaml:"listenAddress"  env:"LISTEN_ADDRESS"  env-default:":9115"`
	ScrapeInterval time.Duration     `yaml:"scrapeInterval" env:"SCRAPE_INTERVAL" env-default:"1h"`
	Output         Output            `yaml:"output"`
	Servers        map[string]Server `yaml:"servers"`
}

// Output selects how metrics are published.
type Output struct {
	Mode         string `yaml:"mode"         env:"OUTPUT_MODE"   env-default:"prometheus"`
	OTLPEndpoint string `yaml:"otlpEndpoint" env:"OTLP_ENDPOINT"`
}

// Server describes one Gitea instance to collect from.
type Server struct {
	URL          string   `yaml:"url"`
	TokenEnvName string   `yaml:"tokenEnvName"`
	ExcludeOrgs  []string `yaml:"excludeOrgs"`
	Token        string   // resolved from TokenEnvName by Load; never serialized
}

// Load reads the config file and environment variables, then resolves
// each server's token from the env var named by its tokenEnvName.
func Load(path string) (Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return Config{}, err
	}
	for name, s := range cfg.Servers {
		if s.TokenEnvName != "" {
			s.Token = os.Getenv(s.TokenEnvName)
			cfg.Servers[name] = s
		}
		if s.Token == "" {
			return Config{}, fmt.Errorf("server %q: no token resolved (check tokenEnvName and that the env var is set)", name)
		}
	}

	return cfg, nil
}

// Validate checks cross-field rules that cleanenv cannot express.
func (c Config) Validate() error {
	switch c.Output.Mode {
	case "prometheus", "otlp":
	default:
		return fmt.Errorf("unknown output mode %q (want \"prometheus\" or \"otlp\")", c.Output.Mode)
	}
	if c.Output.Mode == "otlp" && c.Output.OTLPEndpoint == "" {
		return fmt.Errorf("output.mode=otlp requires output.otlpEndpoint")
	}
	if len(c.Servers) == 0 {
		return fmt.Errorf("no servers configured")
	}
	return nil
}
