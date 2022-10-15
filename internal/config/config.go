package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Logger    LoggerConfig
	DB        DBConfig
	Server    ServerConfig
	Publisher PublisherConfig
}

type LoggerConfig struct {
	Level string
	File  string
}

type DBConfig struct {
	MaxOpenConnections    int
	MaxIdleConnections    int
	MaxConnectionLifetime time.Duration
	DSN                   string
}

type ServerConfig struct {
	Host              string
	Port              int
	ConnectionTimeout time.Duration
}

type PublisherConfig struct {
	URI          string
	QueueName    string
	ExchangeName string
}

func NewConfig(path string) (cfg Config, err error) {
	InitDefaults()
	viper.AutomaticEnv()
	if path != "" {
		viper.SetConfigFile(path)
		viper.SetConfigType("yaml")
		if err := viper.ReadInConfig(); err != nil {
			return Config{}, fmt.Errorf("error during reading config file: %w", err)
		}
	} else {
		fmt.Println("config file path is not specified")
	}

	dbMaxConnectionLifetime, err := time.ParseDuration(viper.GetString("db.maxconnectionlifetime"))
	if err != nil {
		fmt.Printf("config db.maxconnectionlifetime is not correct, default value was used: %s\n", "3m")
		dbMaxConnectionLifetime = time.Minute * 3
	}
	serverConnectionTimeout, err := time.ParseDuration(viper.GetString("server.connectiontimeout"))
	if err != nil {
		fmt.Printf("config server.connectiontimeout is not correct, default value was used: %s\n", "5s")
		serverConnectionTimeout = time.Second * 5
	}

	return Config{
		Logger: LoggerConfig{
			Level: viper.GetString("logger.level"),
			File:  viper.GetString("logger.file"),
		},
		DB: DBConfig{
			MaxOpenConnections:    viper.GetInt("db.maxopenconnections"),
			MaxIdleConnections:    viper.GetInt("db.maxidleconnections"),
			MaxConnectionLifetime: dbMaxConnectionLifetime,
			DSN:                   viper.GetString("db.dsn"),
		},
		Server: ServerConfig{
			Host:              viper.GetString("server.host"),
			Port:              viper.GetInt("server.port"),
			ConnectionTimeout: serverConnectionTimeout,
		},
		Publisher: PublisherConfig{
			URI:          viper.GetString("publisher.uri"),
			QueueName:    viper.GetString("publisher.queuename"),
			ExchangeName: viper.GetString("publisher.exchangename"),
		},
	}, nil
}

func InitDefaults() {
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.file", "./rotation_log.log")
	viper.SetDefault("db.maxopenconnections", 20)
	viper.SetDefault("db.maxidleconnections", 5)
	viper.SetDefault("db.maxconnectionlifetime", "3m")
	viper.SetDefault("db.dsn", "host=localhost user=postgres password=password dbname=rotation sslmode=disable")
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 50051)
	viper.SetDefault("server.connectiontimeout", "5s")
	viper.SetDefault("publisher.uri", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("publisher.queuename", "banner-stats-queue")
	viper.SetDefault("publisher.exchangename", "banner-stats-exchange")
}
