package endpoints

import (
	"github.com/gin-gonic/gin"
	"jgit.me/tools/patroni-web-backend/db"
	"net/http"
)

func AddMemo(c *gin.Context) {

	memo := db.Memo{}

	err := c.ShouldBindJSON(&memo)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	err = memo.Check()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	//TODO: check author throw IPA ?

	err = memo.Create()

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": memo.ID})

}

func UpdateMemo(c *gin.Context) {

	memo := db.Memo{}

	err := c.ShouldBindJSON(&memo)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	err = memo.Check()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	err = memo.Save()

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
}

func DeleteMemo(c *gin.Context) {

	memo := db.Memo{}

	err := c.ShouldBindJSON(&memo)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	err = memo.Delete()

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
}

func Options(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		return
	}
}
