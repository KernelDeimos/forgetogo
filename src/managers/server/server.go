package server

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"time"

	"github.com/dube-dev/forgetogo/src/application"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

//go:embed ui/*
var uiFS embed.FS

type sansRootFS struct {
	Key     string
	Content embed.FS
}

func (rfs sansRootFS) Open(name string) (fs.File, error) {
	return rfs.Content.Open(path.Join(rfs.Key, name))
}

type ServerManager struct{}

func NewServerManager() *ServerManager {
	return &ServerManager{}
}

func SlowMiddleware(c *gin.Context) {
	<-time.After(1000 * time.Millisecond)
	c.Next()
}

func (server *ServerManager) Start(api application.ApplicationAPI) error {
	r := gin.Default()

	// r.Use(SlowMiddleware)

	// TODO: endpoints should be implied by managers

	r.GET("/config/launchers", func(c *gin.Context) {
		result, err := api.RequestSync("config", "launcher configs")
		if err != nil {
			logrus.Error(err)
			c.AbortWithStatusJSON(500, err.Error())
			return
		}
		c.JSON(200, result)
	})

	r.POST("/config/launchers", func(c *gin.Context) {
		newConfig := map[string]interface{}{}
		err := c.BindJSON(&newConfig)
		fmt.Printf("newConfig1: %v\n", newConfig)
		if err != nil {
			logrus.Error("bindJSON", err)
			c.AbortWithStatusJSON(400, err.Error())
			return
		}
		_, err = api.RequestSync(
			"config",
			"create launcher config",
			newConfig,
		)
		if err != nil {
			logrus.Error(err)
			c.AbortWithStatusJSON(500, err.Error())
			return
		}
		c.JSON(200, map[string]interface{}{
			"status": "ok",
		})
	})

	r.GET("/instance/instances", func(c *gin.Context) {
		result, err := api.RequestSync("instance", "get instances")
		if err != nil {
			logrus.Error(err)
			c.AbortWithStatusJSON(500, err.Error())
			return
		}
		c.JSON(200, result)
	})

	r.POST("/instance/instance/:uuid", func(c *gin.Context) {
		uuid := c.Param("uuid")
		result, err := api.RequestSync(
			"instance", "new instance", uuid)

		if err != nil {
			logrus.Error(err)
			c.AbortWithStatusJSON(500, err.Error())
			return
		}
		c.JSON(200, result)
	})

	wsupgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	r.GET("/instance/buffer/:uuid", func(c *gin.Context) {
		uuid := c.Param("uuid")
		logrus.Info("instance buffer handler", uuid)

		result, err := api.RequestSync("instance", "get buffer", uuid)
		if err != nil {
			logrus.Error(err)
			c.AbortWithStatusJSON(500, err.Error())
			return
		}

		c.JSON(200, result)

	})

	r.GET("/instance/stream/:uuid", func(c *gin.Context) {
		uuid := c.Param("uuid")
		logrus.Info("instance stream handler", uuid)
		conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logrus.Fatal(err)
			return
		}

		api.RequestSyncCritical(
			"instance", "attach connection", uuid, conn)
	})

	r.POST("/instance/command/:uuid", func(c *gin.Context) {
		uuid := c.Param("uuid")
		var cmd string
		c.BindJSON(&cmd)
		_, err := api.RequestSync(
			"instance",
			"send command",
			uuid,
			cmd,
		)
		if err != nil {
			logrus.Error(err)
			c.AbortWithStatusJSON(500, err.Error())
			return
		}

		c.JSON(200, map[string]interface{}{
			"status": "ok",
		})
	})

	// r.StaticFS("/", EmbedFolder(uiFS, "/"))
	// r.GET("/*filepath", func(c *gin.Context) {
	r.NoRoute(func(c *gin.Context) {
		c.FileFromFS(c.Request.URL.Path, http.FS(sansRootFS{
			Key:     "ui",
			Content: uiFS,
		}))
	})

	go r.Run(":8085")

	return nil
}
