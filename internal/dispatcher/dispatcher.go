package dispatcher

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"xlab-feishu-robot/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// @Summary feishu event dispatcher
// @Tags feishu_event
// @Accept json
// @Success 200 {string} OK
// @Router /feishu_events [post]
func Dispatcher(c *gin.Context) {
	// Handler for Feishu Event Http Callback

	// [steps]
	// - decrypt if needed
	// - return to test event
	// - check data
	// - dispatch events

	// see: https://open.feishu.cn/document/ukTMukTMukTM/uUTNz4SN1MjL1UzM

	// get raw body (bytes)
	rawBody, _ := ioutil.ReadAll(c.Request.Body)

	// decrypt data if ENCRYPT is on
	var requestStr string
	if encryptKey := config.C.Feishu.EncryptKey; encryptKey != "" {
		rawBodyJson := make(map[string]any)
		json.Unmarshal(rawBody, &rawBodyJson)
		rawRequestStr, _ := rawBodyJson["encrypt"].(string)
		var err error
		requestStr, err = decrypt(rawRequestStr, encryptKey)
		if err != nil {
			logrus.Error("Cannot decrypt request")
		}
	} else {
		requestStr = string(rawBody)
	}

	var req FeishuEventRequest
	deserializeRequest(requestStr, &req)
	logrus.Debug("Feishu Robot received a request: ", req)

	// return to server test event
	if req.Challenge != "" {
		c.JSON(http.StatusOK, gin.H{"challenge": req.Challenge})
		return
	}
}
