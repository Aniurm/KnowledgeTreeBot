package internal

import (
	"xlab-feishu-robot/internal/dispatcher"
	eventHandler "xlab-feishu-robot/internal/event_handler"

	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine) {
	eventHandler.Init()
	Register(r)
}

func Register(r *gin.Engine) {
	r.GET("/api/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	// DO NOT CHANGE LINES BELOW
	// register dispatcher
	r.POST("/feiShu/Event", dispatcher.Dispatcher)
}
