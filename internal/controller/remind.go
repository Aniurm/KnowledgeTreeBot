package controller

import (
	"strings"
	"time"
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

// Record 定义一个结构，用于存储知识树表格中每一个Record的解析结果
type Record struct {
	// 多行文本
	MultiLineText string
	// 维护人
	Maintainers []Maintainer
	// 一句话介绍
	OneLineIntroduction string
	// 维护的节点链接
	NodeLink []Link
	// 创建时间
	TimeStamp float64
	// 👍
	LikeCount int
}

func getAllRecordsInTable(table feishuapi.TableInfo) []Record {
	allRecordData := pkg.Cli.DocumentGetAllRecordsWithLinks(table.AppToken, table.TableId)
	result := make([]Record, 0)
	for _, recordData := range allRecordData {
		result = append(result, parseRecordFields(recordData.Fields))
	}
	return result
}

// Maintainer 定义一个结构，用于存储维护人的信息
type Maintainer struct {
	Name string
	ID   string
}

type Link struct {
	URL         string
	Token       string
	Text        string
	MentionType string
}

// 从API返回的record的Fields中解析出Record信息
// 如果某个字段没写，读取map时会返回nil，所以要检查并处理
func parseRecordFields(record map[string]interface{}) Record {
	result := Record{}
	// 解析多行文本
	if record["多行文本"] != nil {
		result.MultiLineText = parseMultilineText(record["多行文本"])
	}
	// 解析维护人
	if record["维护人"] != nil {
		maintainers := record["维护人"].([]interface{})
		for _, maintainer := range maintainers {
			maintainerMap := maintainer.(map[string]interface{})
			result.Maintainers = append(result.Maintainers, Maintainer{
				Name: maintainerMap["name"].(string),
				ID:   maintainerMap["id"].(string),
			})
		}
	}
	// 解析一句话介绍
	if record["一句话介绍"] != nil {
		result.OneLineIntroduction = parseMultilineText(record["一句话介绍"])
	}
	// 解析维护的节点链接
	if record["维护节点链接"] != nil {
		result.NodeLink = parseLinkFromMultilineText(record["维护节点链接"])
	}
	// 解析创建时间
	if record["创建时间"] != nil {
		result.TimeStamp = record["创建时间"].(float64)
	}
	// 解析点赞数
	if record["👍"] != nil {
		result.LikeCount = int(record["👍"].(float64))
	}

	return result
}

func parseMultilineText(multilineTextData interface{}) string {
	// 多行文本的底层数据是一个数组，数组中的每个元素是一个map，这里一步步解析
	multilineTextMap := multilineTextData.([]interface{})
	var sb strings.Builder
	for _, v := range multilineTextMap {
		currentMap := v.(map[string]interface{})
		currentText := currentMap["text"]
		if currentText != nil {
			sb.WriteString(currentText.(string))
		}
	}
	return sb.String()
}

func parseLinkFromMultilineText(multilineTextData interface{}) []Link {
	// 多行文本的底层数据是一个数组，数组中的每个元素是一个map，这里一步步解析
	multilineTextMap := multilineTextData.([]interface{})
	var result []Link
	for _, v := range multilineTextMap {
		currentMap := v.(map[string]interface{})
		// 筛选出包含link的map
		if _, ok := currentMap["link"]; ok {
			url := currentMap["link"].(string)
			token := currentMap["token"].(string)
			text := currentMap["text"].(string)
			mentionType := currentMap["mentionType"].(string)
			result = append(result, Link{
				URL:         url,
				Token:       token,
				Text:        text,
				MentionType: mentionType,
			})
		}
	}
	return result
}

func parseTimestamp(timestamp float64) (int, int) {
	t := time.Unix(int64(timestamp/1000), 0)
	return t.Year(), int(t.Month())
}

func getTableByTime(year int, month int) feishuapi.TableInfo {
	// 获取所有表格
	allTables := getAllTables()
	for _, table := range allTables {
		// 获取表格中的所有记录
		allRecords := getAllRecordsInTable(table)
		for _, record := range allRecords {
			recordYear, recordMonth := parseTimestamp(record.TimeStamp)
			if recordYear == year && recordMonth == month {
				return table
			}
		}
	}
	return feishuapi.TableInfo{}
}
