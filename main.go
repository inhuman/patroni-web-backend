package main

import (
	"github.com/gin-gonic/gin"
	"jgit.me/tools/patroni-web-backend/config"
	"jgit.me/tools/patroni-web-backend/db"
	"jgit.me/tools/patroni-web-backend/init-app"
	"jgit.me/tools/patroni-web-backend/router"
	"log"
	"os"
)

func main() {

	err := runApp()
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	} else {
		os.Exit(0)
	}

}

func runApp() error {

	err := init_app.Do()
	if err != nil {
		return err
	}

	db.Stor.Migrate(db.ClusterSwitchover{})
	db.Stor.Migrate(db.Memo{})

	r := router.Setup(gin.Logger(), gin.Recovery())
	return r.Run(":" + config.AppConf.Port)
}
