package model

import (
	"strings"
)

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

// ParseRecordFields 从API返回的record的Fields中解析出Record信息
// 如果某个字段没写，读取map时会返回nil，所以要检查并处理
func ParseRecordFields(record map[string]interface{}) Record {
	result := Record{}
	// 解析多行文本
	if record["多行文本"] != nil {
		result.MultiLineText = ParseMultilineText(record["多行文本"])
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
		result.OneLineIntroduction = ParseMultilineText(record["一句话介绍"])
	}
	// 解析维护的节点链接
	if record["维护节点链接"] != nil {
		result.NodeLink = ParseLinkFromMultilineText(record["维护节点链接"])
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

func ParseMultilineText(multilineTextData interface{}) string {
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

func ParseLinkFromMultilineText(multilineTextData interface{}) []Link {
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
