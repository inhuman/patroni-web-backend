package endpoints

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"jgit.me/tools/patroni-web-backend/clusters"
	"jgit.me/tools/patroni-web-backend/config"
	"jgit.me/tools/patroni-web-backend/consul"
	"jgit.me/tools/patroni-web-backend/ipa"
	"jgit.me/tools/patroni-web-backend/utils"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func Status(c *gin.Context) {

	c.JSON(200, gin.H{
		"name":    "patroni web backend",
		"version": config.AppConf.Version,
	})
}

func GetConfig(c *gin.Context) {

	nodesConf := clusters.FetchClusters(consul.Client)

	if nodesConf == nil {
		c.JSON(404, gin.H{"error": "can not fetch any cluster from consul - check consul prefix or clusters config existence in config"})
		c.Abort()
		return
	}

	c.JSON(200, nodesConf)
}

func GetConsulConfig(c *gin.Context) {
	yml := clusters.FetchConsulConfig(consul.Client)
	io.Copy(c.Writer, strings.NewReader(yml))
}

func SaveClusterConsulConfig(c *gin.Context) {

	clusterName := c.Param("cluster")

	yaml, err := c.GetRawData()

	if err != nil {
		fmt.Printf("yaml error: %s\n", err)
		c.AbortWithStatusJSON(400, err)
		return
	}

	err = consul.WriteConfigFromYaml(clusterName, yaml)
	if err != nil {
		fmt.Printf("write config error: %s\n", err)
		c.AbortWithStatusJSON(400, err)
		return
	}

	c.JSON(200, nil)
}

func SetClusterChecksEnabled(c *gin.Context) {
	clusterName := c.Param("cluster")

	err := consul.SetClusterCheck(clusterName, true)
	if err != nil {
		fmt.Printf("write config error: %s\n", err)
		c.AbortWithStatusJSON(400, err)
		return
	}
}

func SetClusterChecksDisabled(c *gin.Context) {
	clusterName := c.Param("cluster")
	err := consul.SetClusterCheck(clusterName, false)
	if err != nil {
		fmt.Printf("write config error: %s\n", err)
		c.AbortWithStatusJSON(400, err)
		return
	}
}

func DeleteClusterConsulConfig(c *gin.Context) {
	clusterName := c.Param("cluster")
	fmt.Println("deleting config for cluster", clusterName)

	err := consul.DeleteConfig(clusterName)
	if err != nil {
		c.JSON(400, err)
		c.Abort()
	}

	c.JSON(200, nil)
}

func GetConfigFull(c *gin.Context) {

	nodesConf := clusters.FetchClusters(consul.Client)
	if nodesConf == nil {
		c.JSON(404, gin.H{"error": "can not fetch any cluster from consul - check consul prefix or clusters config existence in config"})
		c.Abort()
		return
	}

	if err := nodesConf.FetchNodesDetailed(); err != nil {
		c.JSON(405, gin.H{"error": "can not fetch nodes detailed: " + err.Error()})
		c.Abort()
		return
	}

	c.JSON(200, nodesConf)

}

func ReverseProxy() gin.HandlerFunc {

	return func(c *gin.Context) {

		if c.Request.Method == "OPTIONS" {
			return
		}

		uri := c.Param("url")
		uri = "http:/" + uri

		u, err := url.Parse(uri)

		if err != nil {
			fmt.Printf("url parse error: %s\n", err)
		} else {
			fmt.Printf("parced url: %+v\n", u)
		}

		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL = u
			}}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func Ping(c *gin.Context) {

	uri := c.Param("url")
	pinger, err := utils.GetPinger(uri)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	pinger.Count = 3
	pinger.Run()
	stats := pinger.Statistics()
	c.JSON(200, stats)
}

func LoginIPA(c *gin.Context) {
	user := c.PostForm("user")
	password := c.PostForm("password")

	u, err := ipa.AuthUserJwt(user, password)
	if err != nil {
		fmt.Println(err)
		c.JSON(401, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
	c.JSON(200, u)
}
