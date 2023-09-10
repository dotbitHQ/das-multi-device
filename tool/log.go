package tool

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Log(c *gin.Context) (log *logrus.Entry) {
	defaultLog := logrus.WithFields(logrus.Fields{
		"request_id": "",
		"user_ip":    "",
		"user_agent": "",
	})
	if c == nil {
		log = defaultLog
		return
	}
	logger, exists := c.Get("logger")
	if !exists {
		log = defaultLog
		return
	} else {
		log = logger.(*logrus.Entry)
	}

	return
}
