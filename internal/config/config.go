package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Version   int          `mapstructure:"version"`
	Languages []string     `mapstructure:"languages"`
	Ignore    []string     `mapstructure:"ignore"`
	Risk      RiskConfig   `mapstructure:"risk"`
	Impact    ImpactConfig `mapstructure:"impact"`
	Output    OutputConfig `mapstructure:"output"`
	Cache     CacheConfig  `mapstructure:"cache"`
}

type RiskConfig struct {
	Weights        RiskWeights `mapstructure:"weights"`
	BugFixPatterns []string    `mapstructure:"bug_fix_patterns"`
	Since          string      `mapstructure:"since"`
}

type RiskWeights struct {
	BugFixRatio  int `mapstructure:"bug_fix_ratio"`
	AuthorCount  int `mapstructure:"author_count"`
	ChurnRate    int `mapstructure:"churn_rate"`
	StaleAge     int `mapstructure:"stale_age"`
	FanOut       int `mapstructure:"fan_out"`
	TestCoverage int `mapstructure:"test_coverage"`
}

type ImpactConfig struct {
	MaxDepth     int  `mapstructure:"max_depth"`
	IncludeTests bool `mapstructure:"include_tests"`
}

type OutputConfig struct {
	Format string `mapstructure:"format"`
	Color  string `mapstructure:"color"`
}

type CacheConfig struct {
	Dir string `mapstructure:"dir"`
}

func DefaultConfig() *Config {
	return &Config{
		Version:   1,
		Languages: []string{"go"},
		Ignore: []string{
			"vendor/**",
			"**/*_string.go",
			"**/mock_*.go",
		},
		Risk: RiskConfig{
			Weights: RiskWeights{
				BugFixRatio:  25,
				AuthorCount:  15,
				ChurnRate:    15,
				StaleAge:     10,
				FanOut:       20,
				TestCoverage: 15,
			},
			BugFixPatterns: []string{
				`\bfix(ed|es)?\b`,
				`\bbug\b`,
				`\bhotfix\b`,
				`\bpatch\b`,
			},
			Since: "12 months",
		},
		Impact: ImpactConfig{
			MaxDepth:     3,
			IncludeTests: true,
		},
		Output: OutputConfig{
			Format: "tui",
			Color:  "auto",
		},
		Cache: CacheConfig{
			Dir: ".rippl/cache",
		},
	}
}

func Load(moduleRoot, configPath string) (*Config, bool, error) {
	cfg := DefaultConfig()

	path := configPath
	if !filepath.IsAbs(path) {
		path = filepath.Join(moduleRoot, configPath)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, false, nil
	} else if err != nil {
		return nil, false, fmt.Errorf("stat config file: %w", err)
	}

	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, false, fmt.Errorf("read config file: %w", err)
	}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, false, fmt.Errorf("parse config file: %w", err)
	}

	return cfg, true, nil
}
