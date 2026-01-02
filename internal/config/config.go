package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	// Test defaults
	Mode     string `mapstructure:"mode"`
	Seconds  int    `mapstructure:"seconds"`
	Words    int    `mapstructure:"words"`
	Countdown int   `mapstructure:"countdown"`

	// Display
	NoColor bool `mapstructure:"no_color"`
	Wrap    int  `mapstructure:"wrap"`
	Chart   bool `mapstructure:"chart"`

	// Content
	WordsFile  string `mapstructure:"words_file"`
	QuotesFile string `mapstructure:"quotes_file"`
}

var (
	// Initialize cfg to defaults so callers (like cobra flag setup) can safely
	// read defaults before Load() runs.
	cfg        = Default()
	configFile string
)

// Default returns the default configuration
func Default() Config {
	return Config{
		Mode:      "words",
		Seconds:   30,
		Words:     25,
		Countdown: 3,
		NoColor:   false,
		Wrap:      0, // 0 means auto
		Chart:     true,
	}
}

// SetConfigFile sets a custom config file path
func SetConfigFile(path string) {
	configFile = path
}

// Load reads the configuration from file and environment
func Load() error {
	cfg = Default()

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		configDir, err := os.UserConfigDir()
		if err != nil {
			configDir = filepath.Join(os.Getenv("HOME"), ".config")
		}
		viper.AddConfigPath(filepath.Join(configDir, "mtcli"))
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
	}

	viper.SetEnvPrefix("MTCLI")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("mode", cfg.Mode)
	viper.SetDefault("seconds", cfg.Seconds)
	viper.SetDefault("words", cfg.Words)
	viper.SetDefault("countdown", cfg.Countdown)
	viper.SetDefault("no_color", cfg.NoColor)
	viper.SetDefault("wrap", cfg.Wrap)
	viper.SetDefault("chart", cfg.Chart)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
		// Config file not found is OK
	}

	return viper.Unmarshal(&cfg)
}

// Get returns the current configuration
func Get() Config {
	return cfg
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "mtcli"), nil
}

// GetDataDir returns the data directory path (for SQLite DB)
func GetDataDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return configDir, nil
}

