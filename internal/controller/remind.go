package controller

import (
	"github.com/YasyaKarasu/feishuapi"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"xlab-feishu-robot/internal/config"
	"xlab-feishu-robot/internal/pkg"
)

const (
	remindPersonInCharge    = "请及时创建本月的维护记录"
	remindGroupMembersStart = "请及时开始写本月的知识树文档"
)

func Remind() {
	cronTimer := cron.New()
	// Remind the person in charge to create maintenance record
	// Remind group members to start writing knowledge tree documents
	// every month on the 1st at 10:00
	_, err := cronTimer.AddFunc("0 10 1 * *", func() {
		pkg.Cli.MessageSend(feishuapi.UserUserId, config.C.Info.PersonInChargeID, feishuapi.Text, remindPersonInCharge)
		pkg.Cli.MessageSend(feishuapi.GroupChatId, config.C.Info.GroupID, feishuapi.Text, remindGroupMembersStart)
	})
	if err != nil {
		panic(err)
		logrus.Error("Failed to add cron job")
	}
	cronTimer.Start()
}
