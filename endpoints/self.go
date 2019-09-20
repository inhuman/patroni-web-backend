package endpoints

import (
	"github.com/gin-gonic/gin"
	"jgit.me/tools/patroni-web-backend/config"
	"jgit.me/tools/patroni-web-backend/init-app"
	"jgit.me/tools/patroni-web-backend/selfcheck"
)

func SelfCheck(c *gin.Context) {

	var response = gin.H{}

	response["consul"] = selfcheck.ExternalService{
		Uri:    config.AppConf.ConsulAddress,
		Status: selfcheck.CheckConsulHost(),
	}

	response["cluster_config"] = selfcheck.ExternalService{
		Uri:    config.AppConf.ConsulKvPrefix,
		Status: selfcheck.CheckClusterConfig(),
	}

	if config.AppConf.Ipa.Enabled {
		response["free_ipa_host"] = selfcheck.ExternalService{
			Uri:    config.AppConf.Ipa.Host,
			Status: selfcheck.CheckIpa(),
		}
	} else {
		response["free_ipa_host"] = selfcheck.ExternalService{
			Uri:    "",
			Status: "disabled",
		}
	}

	if config.AppConf.ElasticLog {
		response["elastic_log"] = selfcheck.ExternalService{
			Uri:    config.AppConf.ElasticHost,
			Status: selfcheck.CheckElastic(),
		}
	} else {
		response["elastic_log"] = selfcheck.ExternalService{
			Uri:    "",
			Status: "disabled",
		}
	}

	c.JSON(200, response)
}

func Reinit(c *gin.Context) {

	//TODO: save auth data when reinit?

	err := init_app.Do()
	if err != nil {
		c.JSON(500, gin.H{"status": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "ok"})
}
