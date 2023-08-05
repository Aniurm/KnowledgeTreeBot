package chat

import (
	"strings"
	"xlab-feishu-robot/internal/model"

	"github.com/sirupsen/logrus"
)

var groupMessageMap = make(map[string]messageHandler)

func group(messageevent *model.MessageEvent) {
	switch strings.ToUpper(messageevent.Message.Message_type) {
	case "TEXT":
		groupTextMessage(messageevent)
	default:
		logrus.WithFields(logrus.Fields{"message type": messageevent.Message.Message_type}).Warn("Receive group message, but this type is not supported")
	}
}

func groupTextMessage(messageevent *model.MessageEvent) {
	// If the robot is triggered by accident, return
	if isAccident(messageevent) {
		return
	}
	// Remove the prefix and suffix of the message content
	messageevent.Message.Content = strings.TrimSuffix(strings.TrimPrefix(messageevent.Message.Content, "{\"text\":\""), "\"}")
	// Get valid message content
	messageevent.Message.Content = messageevent.Message.Content[strings.Index(messageevent.Message.Content, " ")+1:]
	logrus.WithFields(logrus.Fields{"message content": messageevent.Message.Content}).Info("Receive group TEXT message")

	groupMessageMap["drawlots"](messageevent)
}

func GroupMessageRegister(f messageHandler, s string) {

	if _, isEventExist := groupMessageMap[s]; isEventExist {
		logrus.Warning("Double declaration of group message handler: ", s)
	}
	groupMessageMap[s] = f
}

// isAccident is a function to judge whether the robot is triggered by accident
// If is accident, return true; else return false
func isAccident(messageevent *model.MessageEvent) bool {
	if strings.Contains(messageevent.Message.Content, "@_all") {
		return true
	}
	return false
}
