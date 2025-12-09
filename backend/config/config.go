package config

import "github.com/jinzhu/configor"

type Config struct {
	AppConfig AppConfig `env:"APPCONFIG"`
	DBConfig  DBConfig
}

type AppConfig struct {
	APPName            string `default:"pulse-api"`
	Port               int    `env:"PORT" default:"3000"`
	Version            string `default:"x.x.x" env:"VERSION"`
	Env                string `default:"development" env:"ENV"`
	JWTSecret          string `env:"JWT_SECRET" default:"dev-secret-change-in-production"`
	JWTExpirationHours int    `env:"JWT_EXPIRATION_HOURS" default:"24"`
}

type DBConfig struct {
	Host     string `default:"localhost" env:"DBHOST"`
	DataBase string `default:"pulse" env:"DBNAME"`
	User     string `default:"pulse" env:"DBUSERNAME"`
	Password string `required:"true" env:"DBPASSWORD" default:"mysecretpassword"`
	Port     uint   `default:"5432" env:"DBPORT"`
	SSLMode  string `default:"disable" env:"DBSSL"`
}

func LoadConfigOrPanic() Config {
	var config = Config{}
	configor.Load(&config, "config/config.dev.json")

	return config
}
