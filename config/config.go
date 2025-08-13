package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Context struct {
	BaseURL        string `mapstructure:"base_url"`
	AuthToken      string `mapstructure:"auth_token"`
	Model          string `mapstructure:"model"`
	SmallFastModel string `mapstructure:"small_fast_model"`
}

type Config struct {
	Contexts map[string]Context `mapstructure:"context"`
}

func resolveEnvVar(value string) (string, error) {
	if strings.HasPrefix(value, "env:") {
		envVar := strings.TrimPrefix(value, "env:")
		if envVar == "" {
			return "", fmt.Errorf("environment variable name cannot be empty")
		}
		
		envValue := os.Getenv(envVar)
		if envValue == "" {
			return "", fmt.Errorf("environment variable '%s' is not set or empty", envVar)
		}
		
		return envValue, nil
	}
	return value, nil
}

func GetConfigPath() (string, error) {
	// Check for environment variable override first
	if path := os.Getenv("CCCTX_CONFIG_PATH"); path != "" {
		return path, nil
	}
	
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ccctx", "config.toml"), nil
}

func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Create config directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	// Create default config file if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := `# Claude-Code Context Configuration
[context.example]
base_url = "https://api.anthropic.com"
auth_token = "your-auth-token-here"
# Optional: specify model explicitly
# model = "claude-3-5-sonnet-20241022"
# small_fast_model = "claude-3-5-haiku-20241022"
`
		if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
			return nil, err
		}
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("toml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func ListContexts() ([]string, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	var contexts []string
	for name := range config.Contexts {
		contexts = append(contexts, name)
	}

	return contexts, nil
}

func GetContext(name string) (*Context, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	context, exists := config.Contexts[name]
	if !exists {
		return nil, fmt.Errorf("context '%s' not found", name)
	}

	// Resolve environment variables in auth token
	resolvedAuthToken, err := resolveEnvVar(context.AuthToken)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve auth token for context '%s': %w", name, err)
	}

	resolvedContext := Context{
		BaseURL:        context.BaseURL,
		AuthToken:      resolvedAuthToken,
		Model:          context.Model,
		SmallFastModel: context.SmallFastModel,
	}

	return &resolvedContext, nil
}