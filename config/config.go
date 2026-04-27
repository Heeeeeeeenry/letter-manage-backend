package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server       ServerConfig            `yaml:"server"`
	Database     DatabaseConfig          `yaml:"database"`
	LLM          LLMConfig               `yaml:"llm"`
	Media        MediaConfig             `yaml:"media"`
	CORS         CORSConfig              `yaml:"cors"`
	Environments map[string]EnvOverride  `yaml:"environments"`
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

// EnvOverride 环境覆盖配置，只包含需要按环境区分的字段
type EnvOverride struct {
	Database *DatabaseConfig `yaml:"database"`
	Server   *ServerConfig   `yaml:"server"`
	LLM      *LLMConfig      `yaml:"llm"`
	Media    *MediaConfig    `yaml:"media"`
}

var GlobalConfig *Config

// Load 加载配置：先加载基础配置，再根据 WORK_ENV 合并环境覆盖
func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}

	// 根据 WORK_ENV 合并环境覆盖
	workEnv := os.Getenv("WORK_ENV")
	if workEnv != "" {
		if override, ok := cfg.Environments[workEnv]; ok {
			mergeOverride(cfg, &override)
			log.Printf("WORK_ENV=%s: applied environment overrides", workEnv)
		} else {
			log.Printf("WORK_ENV=%q set but no matching config in environments section", workEnv)
		}
	} else {
		log.Println("WORK_ENV not set, using base config only")
	}

	GlobalConfig = cfg

	// LLM_API_KEY 环境变量覆盖 API Key（优先级最高）
	if envKey := os.Getenv("LLM_API_KEY"); envKey != "" {
		GlobalConfig.LLM.APIKey = envKey
		log.Println("LLM_API_KEY: applied environment override")
	}

	return nil
}

// mergeOverride 将环境覆盖配置合并到主配置
func mergeOverride(dst *Config, src *EnvOverride) {
	if src.Database != nil {
		o := src.Database
		if o.Host != "" {
			dst.Database.Host = o.Host
		}
		if o.Port != 0 {
			dst.Database.Port = o.Port
		}
		if o.Name != "" {
			dst.Database.Name = o.Name
		}
		if o.User != "" {
			dst.Database.User = o.User
		}
		if o.Password != "" {
			dst.Database.Password = o.Password
		}
		if o.Charset != "" {
			dst.Database.Charset = o.Charset
		}
		if o.MaxIdleConns != 0 {
			dst.Database.MaxIdleConns = o.MaxIdleConns
		}
		if o.MaxOpenConns != 0 {
			dst.Database.MaxOpenConns = o.MaxOpenConns
		}
	}
	if src.Server != nil {
		if src.Server.Port != 0 {
			dst.Server.Port = src.Server.Port
		}
		if src.Server.Mode != "" {
			dst.Server.Mode = src.Server.Mode
		}
	}
	if src.LLM != nil {
		if src.LLM.APIURL != "" {
			dst.LLM.APIURL = src.LLM.APIURL
		}
		if src.LLM.APIKey != "" {
			dst.LLM.APIKey = src.LLM.APIKey
		}
		if src.LLM.Model != "" {
			dst.LLM.Model = src.LLM.Model
		}
		if src.LLM.Timeout != 0 {
			dst.LLM.Timeout = src.LLM.Timeout
		}
	}
	if src.Media != nil {
		if src.Media.Root != "" {
			dst.Media.Root = src.Media.Root
		}
	}
}

func Get() *Config {
	return GlobalConfig
}
