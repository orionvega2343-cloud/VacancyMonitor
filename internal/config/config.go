package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	PollInterval      time.Duration     `yaml:"poll_interval"`
	WorkerCount       int               `yaml:"worker_count"`
	ChannelBufferSize int               `yaml:"channel_buffer_size"`
	ShutdownTimeout   int               `yaml:"shutdown_timeout"`
	Filter            FilterHhConfig    `yaml:"filter"`
	RateLimit         RateLimiterConfig `yaml:"rate_limit"`
	Redis             RedisConfig       `yaml:"redis"`
	BotToken          string            `env:"BOT_TOKEN" env-required:"true"`
}

type FilterHhConfig struct {
	Keywords   []string `yaml:"keywords"`
	Experience string   `yaml:"experience"`
	Area       string   `yaml:"area"`
	Remote     bool     `yaml:"remote"`
}

type RateLimiterConfig struct {
	RequestsPerPeriod int `yaml:"requests_per_period"`
	Period            int `yaml:"period"`
}

type RedisConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	DB   int    `yaml:"db"`
}

func MustLoad() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	var cfg Config

	err = cleanenv.ReadConfig("config/config.yml", &cfg)
	if err != nil {
		log.Fatal("Error loading config", err)
	}
	return &cfg

}
