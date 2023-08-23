package logger

import (
	"github.com/sirupsen/logrus"
)

type PrettyFormatter struct {
}

func (PrettyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	time := entry.Time.Format("02-01-2006 15:04:05")

	var lvl string
	switch entry.Level {
	case logrus.PanicLevel:
		lvl = " [PANIC]"
	case logrus.FatalLevel:
		lvl = " [FATAL]"
	case logrus.ErrorLevel:
		lvl = " [ERROR]"
	case logrus.WarnLevel:
		lvl = "  [WARN]"
	case logrus.InfoLevel:
		lvl = "  [INFO]"
	case logrus.DebugLevel:
		lvl = " [DEBUG]"
	case logrus.TraceLevel:
		lvl = " [TRACE]"
	}

	message := entry.Message

	reqCap := len(time) + len(lvl) + len(message) + 4

	log := make([]byte, 0, reqCap)
	log = append(log, time...)
	log = append(log, lvl...)
	log = append(log, " - "...)
	log = append(log, message...)
	log = append(log, '\n')

	return log, nil
}

func Initialize() {
	logrus.SetFormatter(PrettyFormatter{})
}
