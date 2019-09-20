package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-uuid"
	"io"
	"io/ioutil"
	"jgit.me/tools/patroni-web-backend/config"
	"jgit.me/tools/patroni-web-backend/db"
	"jgit.me/tools/patroni-web-backend/ipa"
	"jgit.me/tools/patroni-web-backend/logger"
	"jgit.me/tools/patroni-web-backend/utils"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RequestLogger(c *gin.Context) {
	buf, _ := ioutil.ReadAll(c.Request.Body)
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))

	var allowedPath = regexp.MustCompile("proxy")

	go func() {
		if config.AppConf.ElasticLog {

			uri := c.Param("url")
			uri = "http:/" + uri

			u, err := url.Parse(uri)

			if err != nil {
				fmt.Printf("url parse error: %s\n", err)
			}

			body := readBody(rdr1)

			if utils.StringInSlice(c.Request.Method, []string{"POST", "PUT", "PATCH"}) &&
				allowedPath.MatchString(c.Request.URL.Path) {

				row := logger.Record{
					Time:       time.Now(),
					UserName:   c.GetHeader("Patroni-User"),
					Action:     strings.Trim(u.Path, "/"),
					ActionData: body,
					Url:        uri,
				}

				_, ierr := logger.Client.Index().
					Index("patroni-web-ui-").
					Type("patroni-web-ui").
					BodyJson(row).
					Do(c)

				if ierr != nil {
					log.Printf("elastic insert error: %+v", ierr)
				}
			}
			fmt.Println("Request body log: " + body)
		}
	}()

	c.Request.Body = rdr2
	c.Next()
}

func ScheduledActionsLogger(c *gin.Context) {
	buf, _ := ioutil.ReadAll(c.Request.Body)
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))

	uri := c.Param("url")
	uri = "http:/" + uri

	var switchoverRegexp = regexp.MustCompile("switchover")

	if switchoverRegexp.MatchString(uri) {
		body := readBody(rdr1)

		switchOverRequest := db.ClusterSwitchover{}

		err := json.Unmarshal([]byte(body), &switchOverRequest)
		if err != nil {
			log.Printf("switchover request unmarshal err: %+v", err)

		} else {

			//TODO: find in elastic something like this 'Awaiting failover at 2018-08-14T11:44:00+03:00 (in 3352 seconds)'
			//TODO: for prove of switchover

			Uuid, err := uuid.GenerateUUID()
			if err != nil {
				fmt.Println("Can not generate request uuid")
			}
			c.Header("uuid", Uuid)

			switchOverRequest.ClusterName = c.GetHeader("cluster")
			db.SwitchoverMap[Uuid] = &switchOverRequest
		}
	}

	c.Request.Body = rdr2
	c.Next()
}

func ResponseLogger(c *gin.Context) {
	blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Next()

	db.SaveScheduledSwitchover(c)
	fmt.Println("Response body log: " + blw.body.String())
}

func CheckAuth(c *gin.Context) {
	token := c.GetHeader("Auth-token")

	if err := ipa.CheckAuth(token); err != nil {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}
}

func CheckDcAuth(c *gin.Context) {

	token := c.GetHeader("Auth-token")
	dc := c.GetHeader("Dc")

	u, err := ipa.GetUserByToken(token)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Can not fetch user"})
		return
	}

	fmt.Printf("chekc dc auth user: %+v\n", u)

	if !utils.StringInSlice("patroni-dc-"+dc, u.Groups) {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized dc admin"})
		return
	}
}

func readBody(reader io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	s := buf.String()
	return s
}
