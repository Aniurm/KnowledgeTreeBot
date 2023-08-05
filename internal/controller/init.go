package controller

import (
	"xlab-feishu-robot/internal/chat"
	"xlab-feishu-robot/internal/dispatcher"
)

func InitEvent() {
	dispatcher.RegisterListener(chat.Receive, "im.message.receive_v1")
	InitMessageBind()
}

func InitMessageBind() {
}
