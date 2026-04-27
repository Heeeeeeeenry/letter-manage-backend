package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	LLM      LLMConfig      `yaml:"llm"`
	Media    MediaConfig    `yaml:"media"`
	CORS     CORSConfig     `yaml:"cors"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Name         string `yaml:"name"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Charset      string `yaml:"charset"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
	MaxOpenConns int    `yaml:"max_open_conns"`
}

type LLMConfig struct {
	APIURL  string `yaml:"api_url"`
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
	Timeout int    `yaml:"timeout"`
}

type MediaConfig struct {
	Root string `yaml:"root"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

var GlobalConfig *Config

func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}

	// Override database config based on WORK_ENV environment variable
	workEnv := os.Getenv("WORK_ENV")
	switch workEnv {
	case "home":
		cfg.Database.Host = "127.0.0.1"
		cfg.Database.User = "root"
		cfg.Database.Password = "000000"
		cfg.Database.Name = "letter_manage_db"
		log.Println("WORK_ENV=home: using local database (127.0.0.1, letter_manage_db)")
	case "company":
		cfg.Database.Host = "10.25.65.177"
		cfg.Database.Name = "letter_manage_db"
		log.Println("WORK_ENV=company: using company database (10.25.65.177, letter_manage_db)")
	default:
		log.Printf("WORK_ENV=%q not set or unknown, using config.yaml defaults", workEnv)
	}

	GlobalConfig = cfg
	return nil
}

func Get() *Config {
	return GlobalConfig
}
