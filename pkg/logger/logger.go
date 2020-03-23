package logger

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/seniorGolang/i2s/pkg/logger/format"
)

var Log Logger

type Logger = *logrus.Entry

func init() {
	Log = logrus.WithTime(time.Now())
	logrus.SetFormatter(&format.Formatter{TimestampFormat: time.StampMilli})
}
