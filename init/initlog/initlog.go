package initlog

import (
	"FSRV_Edge/service/logservice"
	"fmt"
	"os"

	go_logger "github.com/phachon/go-logger"
)

// Logger 歷程記錄變數
var logger *go_logger.Logger

func init() {
	if _, err := os.Stat("./EG_log"); os.IsNotExist(err) {
		if err := os.Mkdir("./EG_log", os.ModePerm); err != nil {
			panic(fmt.Sprintf("Make directory error:%s", err.Error()))
		}
	}
	var getLoggerErr error
	logger, getLoggerErr = logservice.GetLogger()
	if getLoggerErr != nil {
		panic(fmt.Sprintf("Get logger error:%s", getLoggerErr.Error()))
	}
}

// GetLogger 取得 logger variable
func GetLogger() *go_logger.Logger {
	return logger
}
