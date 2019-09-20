package selfcheck

import (
	"context"
	"fmt"
	"jgit.me/tools/patroni-web-backend/clusters"
	"jgit.me/tools/patroni-web-backend/config"
	"jgit.me/tools/patroni-web-backend/consul"
	"jgit.me/tools/patroni-web-backend/ipa"
	"jgit.me/tools/patroni-web-backend/logger"
	"jgit.me/tools/patroni-web-backend/utils"
	"strconv"
)

type ExternalService struct {
	Uri    string `json:"uri"`
	Status string `json:"status"`
}

func CheckConsulHost() string {

	_, err := consul.Client.Status().Leader()
	if err != nil {
		return fmt.Sprintf("%s", err)
	}

	return "ok"
}

func CheckClusterConfig() string {

	nodesConf := clusters.FetchClusters(consul.Client)

	if nodesConf == nil {
		return "can not fetch any cluster from consul - check consul prefix or clusters config existence in config"
	}

	return "ok"
}

func CheckIpa() string {

	// works only if logged in
	_, err := ipa.Client.Ping()

	pinger, err := utils.GetPinger(config.AppConf.Ipa.Host)
	if err != nil {
		return err.Error()
	}

	pinger.Count = 2
	pinger.Run()
	l := pinger.Statistics().PacketLoss

	if l > 1 {
		return "Ping packets lost: " + fmt.Sprintf("%f", l)
	}

	return "ok"
}

func CheckElastic() string {

	var ctx = context.Background()

	err := logger.Init()
	if err != nil {
		return err.Error()
	}

	_, statusCode, err := logger.Client.Ping(config.AppConf.ElasticHost).Do(ctx)

	if err != nil {
		return strconv.Itoa(statusCode) + " - " + err.Error()
	}

	return "ok"
}
