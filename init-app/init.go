package init_app

import (
	"fmt"
	"jgit.me/tools/patroni-web-backend/config"
	"jgit.me/tools/patroni-web-backend/consul"
	"jgit.me/tools/patroni-web-backend/ipa"
	"jgit.me/tools/patroni-web-backend/logger"
)

func Do() error {

	err := config.AppConf.Load()
	if err != nil {
		return err
	}

	if config.AppConf.ElasticLog {
		err := logger.Init()
		if err != nil {
			fmt.Println(err)
		}
	}

	if config.AppConf.Ipa.Enabled {
		ipa.Init()
	}

	err = consul.Init()
	if err != nil {
		fmt.Println(err)
	}

	return nil
}
