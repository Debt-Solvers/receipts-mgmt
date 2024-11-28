package configs

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
		SSLMode  string `mapstructure:"sslmode"`
	} `mapstructure:"database"`
	JWT struct {
		Secret           string `mapstructure:"secret"`
		ExpirationHours  int    `mapstructure:"expiration_hours"`
	} `mapstructure:"jwt"`
}

type AzureConfig struct {
	ComputerVision struct {
		Key      string `mapstructure:"key"`
		Endpoint string `mapstructure:"endpoint"`
	} `mapstructure:"computer_vision"`
	CustomVision struct {
		Key        string `mapstructure:"key"`
		Endpoint   string `mapstructure:"endpoint"`
		ProjectID  string `mapstructure:"project_id"`
		Iteration 	string `mapstructure:"iteration"`
		URL 				string `mapstructure:"url"`
	} `mapstructure:"custom_vision"`
	DocumentIntelligence struct {
		Key      string `mapstructure:"key"`
		Endpoint string `mapstructure:"endpoint"`
	} `mapstructure:"document_intelligence"`
	
}

// LoadConfig reads configuration from file and environment variables
func LoadConfig() *Config {
	var config Config

	// Set up Viper
	viper.SetConfigName("config")         // name of config file (without extension)
	viper.SetConfigType("yaml")           // if the config file is a YAML file
	viper.AddConfigPath("./configs")      // path to look for the config file

	// Enable ENVIRONMENT VARIABLES
	viper.AutomaticEnv()                  // automatically override config with env vars
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // replace '.' with '_' in env vars

	// Read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Error reading config file: %s", err)
		// Continue execution as env vars might be set
	}

	// Configure mappings for nested values with DB_ prefix
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.name", "DB_NAME")
	viper.BindEnv("database.sslmode", "DB_SSLMODE")
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("jwt.expiration_hours", "JWT_EXPIRATION_HOURS")

	// Add Azure bindings
	viper.BindEnv("azure.computer_vision.key", "AZURE_COMPUTER_VISION_KEY")
	viper.BindEnv("azure.computer_vision.endpoint", "AZURE_COMPUTER_VISION_ENDPOINT")
	viper.BindEnv("azure.custom_vision.key", "AZURE_CUSTOM_VISION_KEY")
	viper.BindEnv("azure.custom_vision.endpoint", "AZURE_CUSTOM_VISION_ENDPOINT")
	viper.BindEnv("azure.custom_vision.project_id", "AZURE_CUSTOM_VISION_PROJECT_ID")
	viper.BindEnv("azure.custom_vision.iteration", "AZURE_CUSTOM_VISION_ITERATION")
	viper.BindEnv("azure.custom_vision.url", "AZURE_CUSTOM_VISION_URL")
	viper.BindEnv("azure.document_intelligence.key", "AZURE_DOCUMENT_INTELLIGENCE_KEY")
	viper.BindEnv("azure.document_intelligence.endpoint", "AZURE_DOCUMENT_INTELLIGENCE_ENDPOINT")

	// Unmarshal the configuration into struct
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}

	return &config
}


