// logger/logger.go

package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var BaseLogger = logrus.New()

type CustomFormatter struct {
	logrus.TextFormatter
}

func init() {
	BaseLogger.Out = os.Stdout
}
