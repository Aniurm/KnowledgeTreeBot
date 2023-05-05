package controller

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"xlab-feishu-robot/internal/config"
	"xlab-feishu-robot/internal/pkg"

	"github.com/YasyaKarasu/feishuapi"
	"github.com/sirupsen/logrus"
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
		personsNotWritten := getPersonsNotWritten()
		// TODO: send message to remind persons who have not written the knowledge tree document

	})

	cronTimer.Start()
}

// getPersonsNotWritten gets persons who have not written the knowledge tree document
func getPersonsNotWritten() []feishuapi.GroupMember {
	result := make([]feishuapi.GroupMember, 0)
	allMembers := pkg.Cli.GroupGetMembers(config.C.Info.GroupID, feishuapi.OpenId)
	for _, member := range allMembers {
		if _, ok := getPersonWritten()[member.MemberId]; !ok {
			result = append(result, member)
		}
	}
	return result
}

// getPersonWritten get the persons who has written the knowledge tree document, store in a map
func getPersonWritten() (result map[string]bool) {
	allRecords := getLatestRecords()
	for _, record := range allRecords {
		// Check if the field value is a slice of interfaces
		if fieldSlice, ok := record.Fields["维护人"].([]interface{}); ok {
			// Create a new slice to hold the map[string]interface{} values
			maintainers := make([]map[string]interface{}, len(fieldSlice))

			// Type assert each element to map[string]interface{} and add to the new slice
			for i, elem := range fieldSlice {
				if maintainerMap, ok := elem.(map[string]interface{}); ok {
					maintainers[i] = maintainerMap
				} else {
					logrus.Error("Expected map[string]interface{} but found " + fmt.Sprintf("%T", elem))
				}
			}

			for _, maintainer := range maintainers {
				id := maintainer["id"].(string)
				result[id] = true
			}
		} else {
			logrus.Error("Expected []interface{} but found " + fmt.Sprintf("%T", record.Fields["维护人"]))
		}
	}

	return result
}

func getLatestRecords() []feishuapi.RecordInfo {
	bitable := pkg.Cli.DocumentGetAllBitables(getKnowledgeTreeDocumentID())[0]
	table := pkg.Cli.DocumentGetAllTables(bitable.AppToken)[0]
	return pkg.Cli.DocumentGetAllRecords(table.AppToken, table.TableId)
}

func getKnowledgeTreeDocumentID() string {
	logrus.Info("Node token: ", config.C.Info.NodeToken)
	nodeInfo := pkg.Cli.KnowledgeSpaceGetNodeInfo(config.C.Info.NodeToken)
	return nodeInfo.ObjToken
}
