package config

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"log"
	"os"
)

type Config struct {
	Db   DBConfig
	Env  string
	Nats NatsConfig
}

type DBConfig struct {
	Name                string
	LocalAddress        string
	DockerServerAddress string
}

type NatsConfig struct {
	QueueGroup                         string
	SubscriberSubjectSendAlarmDigest   string
	SubscriberSubjectAlarmStatusChange string
	ProducerSubjectAlarmDigest         string
	ServerAddress                      string
	DockerServerAddress                string
}

func (c Config) Build() Config {
	config := loadFrom("application-config.json")

	env, envIsPresent := os.LookupEnv("ENV")
	if !envIsPresent || env == "" {
		env = "dev"
	}

	config.Env = env
	log.Println("Running with env:", env)

	setNatsAddress(&config)

	return config
}

func setNatsAddress(config *Config) {
	switch config.Env {
	case "test":
		log.Printf("Config - setting nats server address to docker address %s", config.Nats.DockerServerAddress)
		config.Nats.ServerAddress = config.Nats.DockerServerAddress
	default:
		config.Nats.ServerAddress = nats.DefaultURL
	}
}

func (d DBConfig) GetAddress(config Config) string {
	switch config.Env {
	case "test":
		return config.Db.DockerServerAddress
	default:
		return config.Db.LocalAddress
	}
}

func loadFrom(path string) Config {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal("Cannot open config file: ", err)
	}
	defer file.Close()

	configuration := Config{}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	if err != nil {
		log.Fatal("Cannot decode json config: ", err)
	}

	return configuration
}
