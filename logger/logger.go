package logger

import (
	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v5"
	"jgit.me/tools/patroni-web-backend/config"
	"time"
)

var Client *elastic.Client

func Init() error {

	var err error

	Client, err = elastic.NewClient(elastic.SetURL(config.AppConf.ElasticHost))
	if err != nil {
		return err
	}

	if !Client.IsRunning() {
		return errors.New("Could not make connection to elastic host, not running")
	}
	return nil
}

type Record struct {
	Time       time.Time   `json:"@timestamp"`
	Message    string      `json:"message"`
	UserName   string      `json:"user_name"`
	Action     string      `json:"action"`
	ActionData interface{} `json:"action_data"`
	Url        string      `json:"url"`
}
