package config

import (
	"errors"
	"github.com/inhuman/config_merger"
	"github.com/joho/godotenv"
	"log"
)

// AppConf is main app config
var AppConf = &backendConfig{}

type backendConfig struct {
	Port           string `env:"PB_PORT" required:"true"`
	Version        string `env:"PB_VERSION"`
	WebURL         string `env:"WEB_UI_URL" required:"true"`
	ConsulAddress  string `env:"CONSUL_HTTP_ADDR" required:"true"`
	ConsulDc       string `env:"CONSUL_DC" required:"true"`
	ConsulKvPrefix string `env:"CONSUL_KV" required:"true"`
	ElasticHost    string `env:"ELASTIC_HOST"`
	ElasticLog     bool   `env:"ELASTIC_LOG"`
	Postgres       postgreConf
	Ipa            IpaAuthConfig
}

type IpaAuthConfig struct {
	Enabled bool   `env:"IPA_AUTH"`
	Host    string `env:"IPA_HOST"`
	Salt    string `env:"IPA_JWT_SALT"`
}

type postgreConf struct {
	Host     string `env:"PGSQL_HOST" required:"true"`
	Port     string `env:"PGSQL_PORT" required:"true"`
	User     string `env:"PGSQL_USER" required:"true"`
	Password string `env:"PGSQL_PASS" required:"true" show_last_symbols:"4"`
	DbName   string `env:"PGSQL_DB" required:"true"`
}

func (c *backendConfig) Load() error {

	err := godotenv.Overload()
	if err != nil {
		log.Println("Fetching environment variables from OS")
	} else {
		log.Println("Fetching environment variables from .env file")
	}

	configMerger := config_merger.NewMerger(c)

	configMerger.AddSource(&config_merger.EnvSource{
		Variables: []string{
			"PB_PORT",
			"PB_VERSION",
			"WEB_UI_URL",
			"CONSUL_HTTP_ADDR",
			"CONSUL_DC",
			"CONSUL_KV",
			"PGSQL_HOST",
			"PGSQL_PORT",
			"PGSQL_USER",
			"PGSQL_PASS",
			"PGSQL_DB",
			"ELASTIC_HOST",
			"ELASTIC_LOG",
			"IPA_AUTH",
			"IPA_HOST",
		},
	})

	err = configMerger.Run()

	if err != nil {
		return err
	}

	configMerger.PrintConfig()

	if (c.ElasticLog == true) && (c.ElasticHost == "") {
		log.Print("Elastic host not found in ELASTIC_HOST, but elastic logging is enabled, exiting...")
		return errors.New("elastic host not found in env")
	}

	if (c.Ipa.Enabled == true) && (c.Ipa.Host == "") {
		log.Print("FreeIPA host not found in IPA_HOST, but freeIPA auth is enabled, exiting...")
		return errors.New("FreeIPA host not found in env")
	}

	return nil
}
