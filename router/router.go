package router

import (
	"github.com/gin-gonic/gin"
	"jgit.me/tools/patroni-web-backend/config"
	"jgit.me/tools/patroni-web-backend/endpoints"
	"jgit.me/tools/patroni-web-backend/middleware"
	"net/http"
)

func Setup(h ...gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Use(h...)

	r.Use(AddCORS, Options)
	r.Use(middleware.ScheduledActionsLogger, middleware.ResponseLogger, middleware.RequestLogger)

	r.GET("/config", endpoints.GetConfig)
	r.GET("/config/consul", endpoints.GetConsulConfig)
	r.GET("/config/full", endpoints.GetConfigFull)

	r.POST("/login/ipa", endpoints.LoginIPA)

	r.GET("/proxy/*url", endpoints.ReverseProxy())
	r.HEAD("/proxy/*url", endpoints.ReverseProxy())
	r.OPTIONS("/proxy/*url", endpoints.ReverseProxy())

	r.GET("/ping/*url", endpoints.Ping)
	r.GET("/self/check", endpoints.SelfCheck)

	r.GET("/", endpoints.Status)

	authorized := r.Group("/")
	authorized.Use(middleware.CheckAuth)
	{
		dcAuth := authorized.Group("/")
		dcAuth.Use(middleware.CheckDcAuth)

		dcAuth.POST("/cluster/:cluster/checks-enabled", endpoints.SetClusterChecksEnabled)
		dcAuth.DELETE("/cluster/:cluster/checks-enabled", endpoints.SetClusterChecksDisabled)

		dcAuth.DELETE("/config/consul/:cluster", endpoints.DeleteClusterConsulConfig)

		dcAuth.POST("/proxy/*url", endpoints.ReverseProxy())
		dcAuth.PUT("/proxy/*url", endpoints.ReverseProxy())
		dcAuth.PATCH("/proxy/*url", endpoints.ReverseProxy())
		dcAuth.DELETE("/proxy/*url", endpoints.ReverseProxy())

		authorized.POST("/config/consul/:cluster", endpoints.SaveClusterConsulConfig)
		authorized.POST("/memo", endpoints.AddMemo)
		authorized.PUT("/memo", endpoints.UpdateMemo)
		authorized.DELETE("/memo", endpoints.DeleteMemo)

		authorized.GET("/self/reinit", endpoints.Reinit)
	}

	return r
}

func AddCORS(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", config.AppConf.WebURL)
	c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE, PATCH")
	c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Auth-token, Cluster, Dc")
}

func Options(c *gin.Context) {
	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		c.Header("Access-Control-Allow-Origin", config.AppConf.WebURL)
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE, PATCH")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Auth-token, Cluster, Dc")
		c.Header("Allow", "HEAD,GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.AbortWithStatus(http.StatusOK)
	}
}
