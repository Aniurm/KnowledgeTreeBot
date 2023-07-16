package controller

import (
	"fmt"
	"strings"
	"xlab-feishu-robot/internal/config"
	"xlab-feishu-robot/internal/pkg"

	"github.com/robfig/cron/v3"

	"github.com/YasyaKarasu/feishuapi"
	"github.com/sirupsen/logrus"
)

const (
	remindPersonInChargeString    = "请及时创建本月的维护记录"
	remindGroupMembersStartString = "请及时开始写本月的知识树文档"
)

func Remind() {
	cronTimer := cron.New()
	// Remind the person in charge to create maintenance record
	// Remind group members to start writing knowledge tree documents
	// every month on the 1st at 10:00
	_, err := cronTimer.AddFunc("0 10 1 * *", func() {
		remindFirstDay()
	})
	if err != nil {
		logrus.Error("Failed to add cron job")
		panic(err)
	}
	logrus.Info("Added cron job on the 1st of every month at 10:00")

	// Every 15th/23rd of the month at 10:00, check who has not written the knowledge tree document
	_, err = cronTimer.AddFunc("0 10 15,23 * *", func() {
		sendRemindMessage()
	})
	if err != nil {
		logrus.Error("Failed to add cron job")
		panic(err)
	}

	// Every 1st of the month at 0:00, send last month's monthly report
	_, err = cronTimer.AddFunc("@monthly", func() {
		sendMonthlyReport()
	})
	if err != nil {
		logrus.Error("Failed to add cron job")
		panic(err)
	}

	logrus.Info("Add jobs successfully, going to start cron timer")
	cronTimer.Start()
}

func sendToGroup(str string) {
	pkg.Cli.MessageSend(feishuapi.GroupChatId, config.C.Info.GroupID, feishuapi.Text, str)
}

func remindFirstDay() {
	pkg.Cli.MessageSend(feishuapi.UserOpenId, config.C.Info.PersonInChargeID, feishuapi.Text, remindPersonInChargeString)
	sendToGroup(remindGroupMembersStartString)
}

func remindNotWritten(personsNotWritten []feishuapi.GroupMember) {
	var sb strings.Builder
	sb.WriteString("滴滴！查询知识树进度：\n")
	for _, person := range personsNotWritten {
		// @ person in the format of <at user_id="xxx">xxx</at>
		sb.WriteString("<at user_id=\"" + person.MemberId + "\">" + person.Name + "</at>")
	}
	sb.WriteString(" \n知识树维护链接：")
	sb.WriteString(config.C.Info.KnowledgeTreeURL)
	logrus.Info("Remind message: ", sb.String())
	sendToGroup(sb.String())
}

// sendMonthlyReport sends monthly report
func sendMonthlyReport() {
	// Get the persons who did not write the knowledge tree document
	personsNotWritten := getPersonsNotWritten()
	if len(personsNotWritten) > 0 {
		reportNotWritten(personsNotWritten)
	} else {
		reportAllWritten()
	}
}

// reportNotWritten sends monthly report when some group members have not written the knowledge tree document
func reportNotWritten(personsNotWritten []feishuapi.GroupMember) {
	var sb strings.Builder
	sb.WriteString("滴滴！本月未完成知识树的同学：\n")
	for _, person := range personsNotWritten {
		// @ person in the format of <at user_id="xxx">xxx</at>
		sb.WriteString("<at user_id=\"" + person.MemberId + "\">" + person.Name + "</at>")
	}
	logrus.Info("Monthly report: ", sb.String())
	sendToGroup(sb.String())
}

// reportAllWritten sends monthly report when all group members have written the knowledge tree document
func reportAllWritten() {
	var sb strings.Builder
	sb.WriteString("滴滴！本月知识树文档已全部完成。\n")
	sendToGroup(sb.String())
}

func sendRemindMessage() {
	personsNotWritten := getPersonsNotWritten()
	if len(personsNotWritten) > 0 {
		remindNotWritten(personsNotWritten)
	} else {
		logrus.Info("All group members have written the knowledge tree document")
	}
}

// getPersonsNotWritten gets persons who have not written the knowledge tree document
func getPersonsNotWritten() []feishuapi.GroupMember {
	result := make([]feishuapi.GroupMember, 0)
	allMembers := pkg.Cli.GroupGetMembers(config.C.Info.GroupID, feishuapi.OpenId)

	personsWritten := getPersonWritten()
	for _, member := range allMembers {
		if _, ok := personsWritten[member.MemberId]; !ok && !isInWhiteList(member.MemberId) {
			// If the member is not in the white list and has not written the knowledge tree document
			// Add the member to the result
			result = append(result, member)
		}
	}
	logrus.Info("Persons who have not written the knowledge tree document: ", result)
	return result
}

// getPersonWritten get the persons who have written the knowledge tree document, store in a map
func getPersonWritten() map[string]bool {
	result := make(map[string]bool)
	allRecords := getLatestRecords()
	logrus.Info("All records: ", allRecords)
	for _, record := range allRecords {
		// Check if the field value is a slice of interfaces
		if fieldSlice, ok := record.Fields["维护人"].([]interface{}); ok {
			// Create a new slice to hold the map[string]interface{} values
			// Get the maintainers of the record
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
				// Check whether "维护节点链接" is nil
				if record.Fields["维护节点链接"] != nil {
					id := maintainer["id"].(string)
					result[id] = true
				}
			}
		} else {
			logrus.Error("Expected []interface{} but found " + fmt.Sprintf("%T", record.Fields["维护人"]))
		}
	}
	logrus.Info("Persons who have written the knowledge tree document: ", result)
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

// 判断是否在白名单中
func isInWhiteList(person string) bool {
	for _, p := range config.C.WhiteList {
		if p == person {
			return true
		}
	}
	return false
}
