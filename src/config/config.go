package config

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"log"
	"os"
	"strings"
)

type Config struct {
	Db  DBConfig
	Env string
	Nats NatsConfig
}

type DBConfig struct {
	Name string
	LocalAddress string
	Address   string
	Password  string
}

type NatsConfig struct {
	QueueGroup                         string
	SubscriberSubjectSendAlarmDigest   string
	SubscriberSubjectAlarmStatusChange string
	ProducerSubjectAlarmDigest         string
	ServerAddress                      string
}

func BuildConfig() Config {
	config := loadFrom("application-config.json")

	env, envIsPresent := os.LookupEnv("ENV")
	password, dbPasswordIsPresent := os.LookupEnv("ALARM_DIGEST_DB_PASSWORD")
	if !envIsPresent || env == "" {
		env = "dev"
	}
	if !dbPasswordIsPresent && env != "dev" {
		log.Fatal("Cannot connect to db as env: ALARM_DIGEST_DB_PASSWORD not found")
	}

	config.Env = env
	log.Println("Running with env:", env)
	config.Db.Password = password
	config.Db.Address = strings.Replace(config.Db.Address, "<password>", password, 1)

	if config.Nats.ServerAddress == "" {
		config.Nats.ServerAddress = nats.DefaultURL
	}

	return config
}

func loadFrom(path string) Config {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal("Cannot open config file: ", err)
	}
	defer file.Close()

	configuration := Config {}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	if err != nil {
		log.Fatal("Cannot decode json config: ", err)
	}

	return configuration
}
