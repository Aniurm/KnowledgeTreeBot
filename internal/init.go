package internal

import (
	eventHandler "xlab-feishu-robot/internal/event_handler"

	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine) {
	eventHandler.Init()
}
