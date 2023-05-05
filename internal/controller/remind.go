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
		logrus.Error("Failed to add cron job")
		panic(err)
	}
	logrus.Info("Added cron job on the 1st of every month at 10:00")

	// Every 15th of the month at 10:00, check who has not written the knowledge tree document
	_, err = cronTimer.AddFunc("0 10 15 * *", func() {
		// TODO: get the user ID of persons who has not written the knowledge tree document
	})

	cronTimer.Start()
}

// getPersonsNotWritten gets the user ID of persons who has not written the knowledge tree document
func getPersonsNotWritten() {
	return
}

// getIDOfPersonWritten get the ID of persons who has written the knowledge tree document
func getIDOfPersonWritten() []string {

}

func getLatestRecord() []feishuapi.RecordInfo {
	bitable := pkg.Cli.DocumentGetAllBitables(getKnowledgeTreeDocumentID())[0]
	table := pkg.Cli.DocumentGetAllTables(bitable.AppToken)[0]
	return pkg.Cli.DocumentGetAllRecords(table.AppToken, table.TableId)
}

func getKnowledgeTreeDocumentID() string {
	logrus.Info("Node token: ", config.C.Info.NodeToken)
	nodeInfo := pkg.Cli.KnowledgeSpaceGetNodeInfo(config.C.Info.NodeToken)
	return nodeInfo.ObjToken
}
