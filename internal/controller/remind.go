package controller

import (
	"strings"
	"xlab-feishu-robot/internal/config"
	"xlab-feishu-robot/internal/model"
	"xlab-feishu-robot/internal/pkg"
	"xlab-feishu-robot/internal/util"

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
	allRecords := getAllRecordsInTable(getLatestTable())
	for _, record := range allRecords {
		// 该记录的维护节点链接必须非空，否则不算写了知识树
		if record.NodeLink != nil {
			if record.Maintainers != nil {
				// 一个record可能有多个维护者
				for _, maintainer := range record.Maintainers {
					result[maintainer.ID] = true
				}
			}
		}
	}
	logrus.Info("Persons who have written the knowledge tree document: ", result)
	return result
}

func getAllTables() []feishuapi.TableInfo {
	// 注意：DocumentGetAllBitables返回的数组中的所有bitable.AppToken是一样的
	// 所以这里直接取第一个bitable的AppToken
	bitable := pkg.Cli.DocumentGetAllBitables(getKnowledgeTreeDocumentID())[0]
	// bitable里面的所有table相当于知识树文档中的所有表格
	return pkg.Cli.DocumentGetAllTables(bitable.AppToken)
}

func getLatestTable() feishuapi.TableInfo {
	// 最新表格在数组的第一个位置
	return getAllTables()[0]
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

func getAllRecordsInTable(table feishuapi.TableInfo) []model.Record {
	allRecordData := pkg.Cli.DocumentGetAllRecordsWithLinks(table.AppToken, table.TableId)
	result := make([]model.Record, 0)
	for _, recordData := range allRecordData {
		result = append(result, model.ParseRecordFields(recordData.Fields))
	}
	return result
}

func getTableByTime(year int, month int) feishuapi.TableInfo {
	// 获取所有表格
	allTables := getAllTables()
	for _, table := range allTables {
		// 获取表格中的所有记录
		allRecords := getAllRecordsInTable(table)
		for _, record := range allRecords {
			recordYear, recordMonth := util.ParseTimestamp(record.TimeStamp)
			if recordYear == year && recordMonth == month {
				return table
			}
		}
	}
	return feishuapi.TableInfo{}
}
