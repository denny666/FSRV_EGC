package logservice

import (
	"errors"

	go_logger "github.com/phachon/go-logger"
)

// GetLogger logger setting
func GetLogger() (logger *go_logger.Logger, err error) {
	defer func() {
		if r := recover(); r != nil {
			// find out exactly what the error was and set err
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	logger = go_logger.NewLogger()
	detachErr := logger.Detach("console")
	if detachErr != nil {
		panic(detachErr)
	}
	consoleConfig := &go_logger.ConsoleConfig{
		Color:  true,
		Format: "%timestamp_format% %body%",
	}
	attachConsoleErr := logger.Attach("console", go_logger.LOGGER_LEVEL_DEBUG, consoleConfig)
	if attachConsoleErr != nil {
		panic(attachConsoleErr)
	}
	fileConfig := &go_logger.FileConfig{
		Filename:   "./EG_log/EG.log",
		MaxSize:    10 * 1024,
		MaxLine:    10000,
		DateSlice:  "d",
		JsonFormat: false,
		Format:     "%millisecond_format% [%level_string%] [%file%:%line%] %body%",
	}
	attachFileErr := logger.Attach("file", go_logger.LOGGER_LEVEL_DEBUG, fileConfig)
	if attachFileErr != nil {
		panic(attachFileErr)
	}

	logger.SetAsync()
	return
}
