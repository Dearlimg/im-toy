package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

type ServerConfig struct {
	Port         int `mapstructure:"port"`
	ReadTimeout  int `mapstructure:"read_timeout"`
	WriteTimeout int `mapstructure:"write_timeout"`
	MaxHeader    int `mapstructure:"max_header"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Charset  string `mapstructure:"charset"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`
}

type JWTConfig struct {
	SecretKey string `mapstructure:"secret_key"`
	Expire    int    `mapstructure:"expire"`
	Refresh   int    `mapstructure:"refresh"`
}

var AppConfig *Config

func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	AppConfig = config
	log.Printf("Config loaded successfully from %s", configPath)
	return config, nil
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		d.Username,
		d.Password,
		d.Host,
		d.Port,
		d.Database,
		d.Charset,
	)
}

func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}
