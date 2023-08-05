package model

type Privileges int

const (
	ProductManagerGroupMembers Privileges = 0
	ProjectGroupLeader         Privileges = 1
	Other                      Privileges = 2
)

// for more detailed information, see https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/events/receive
type MessageEvent struct {
	Sender struct {
		Sender_id struct {
			Union_id string `json:"union_id"`
			Open_id  string `json:"open_id"`
			User_id  string `json:"user_id"`
		} `json:"sender_id"`
		Sender_type string `json:"sender_type"`
		Tenant_key  string `json:"tenant_key"`
	} `json:"sender"`
	Message struct {
		Message_id   string `json:"message_id"`
		Root_id      string `json:"root_id"`
		Parent_id    string `json:"parent_id"`
		Create_time  string `json:"create_time"`
		Chat_id      string `json:"chat_id"`
		Chat_type    string `json:"chat_type"`
		Message_type string `json:"message_type"`
		Content      string `json:"content"`
		Mentions     []struct {
			Key string `json:"key"`
			Id  struct {
				Union_id string `json:"union_id"`
				Open_id  string `json:"open_id"`
				User_id  string `json:"user_id"`
			}
			Name       string `json:"name"`
			Tenant_key string `json:"tenant_key"`
		} `json:"mentions"`
	} `json:"message"`
}
