package db

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"time"
)

type ClusterSwitchover struct {
	gorm.Model
	Leader      string    `gorm:"not null" json:"leader"`
	Candidate   string    `gorm:"not null" json:"candidate"`
	ClusterName string    `gorm:"not null"`
	ScheduledAt time.Time `json:"scheduled_at"`
}

var SwitchoverMap = make(map[string]*ClusterSwitchover)

func SaveScheduledSwitchover(c *gin.Context) {

	switchOverRequestUuid := c.Writer.Header().Get("uuid")

	if swr, ok := SwitchoverMap[switchOverRequestUuid]; ok {

		if (c.Writer.Status() == 202) || (c.Writer.Status() == 200) {
			fmt.Println("Scheduled switchover fine, saving to db")
			Stor.Db().Save(&swr)
			delete(SwitchoverMap, switchOverRequestUuid)

		} else {
			fmt.Println("Scheduled switchover request fail")
		}
	}
}

func (cs *ClusterSwitchover) GetLastScheduledSwitchover() {
	Stor.Db().Where("cluster_name = ?", cs.ClusterName).Last(cs)
}
