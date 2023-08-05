package chat

import (
	"encoding/json"
	"xlab-feishu-robot/internal/model"

	"github.com/sirupsen/logrus"
)

type messageHandler func(event *model.MessageEvent)

// dispatch message, according to Chat type
func Receive(event map[string]any) {
	messageevent := model.MessageEvent{}
	map2struct(event, &messageevent)
	switch messageevent.Message.Chat_type {
	case "group":
		group(&messageevent)
	default:
		logrus.WithFields(logrus.Fields{"chat type": messageevent.Message.Chat_type}).Warn("Receive message, but this chat type is not supported")
	}
}

func map2struct(m map[string]interface{}, stru interface{}) {
	bytes, _ := json.Marshal(m)
	json.Unmarshal(bytes, stru)
}
