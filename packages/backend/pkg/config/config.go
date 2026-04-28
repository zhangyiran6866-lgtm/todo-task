package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 应用全局配置
type Config struct {
	App     AppConfig     `mapstructure:"app"`
	MongoDB MongoDBConfig `mapstructure:"mongodb"`
	JWT     JWTConfig     `mapstructure:"jwt"`
	Log     LogConfig     `mapstructure:"log"`
}

type AppConfig struct {
	Port int    `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

type MongoDBConfig struct {
	URI                     string `mapstructure:"uri"`
	Database                string `mapstructure:"database"`
	MaxPoolSize             uint64 `mapstructure:"max_pool_size"`
	MinPoolSize             uint64 `mapstructure:"min_pool_size"`
	ConnectTimeoutSeconds   int    `mapstructure:"connect_timeout_seconds"`
	OperationTimeoutSeconds int    `mapstructure:"operation_timeout_seconds"`
}

type JWTConfig struct {
	AccessSecret        string `mapstructure:"access_secret"`
	RefreshSecret       string `mapstructure:"refresh_secret"`
	AccessExpireMinutes int    `mapstructure:"access_expire_minutes"`
	RefreshExpireDays   int    `mapstructure:"refresh_expire_days"`
}

type LogConfig struct {
	Level         string `mapstructure:"level"`
	Format        string `mapstructure:"format"`
	AppPath       string `mapstructure:"app_path"`
	ErrorPath     string `mapstructure:"error_path"`
	AuditPath     string `mapstructure:"audit_path"`
	RetentionDays int    `mapstructure:"retention_days"`
	Compress      bool   `mapstructure:"compress"`
	Stdout        bool   `mapstructure:"stdout"`
}

// Load 从指定路径加载配置文件
func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config.Load: read config failed: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config.Load: unmarshal config failed: %w", err)
	}

	return &cfg, nil
}
