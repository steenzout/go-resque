package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	gol "github.com/steenzout/go-log"
	"github.com/steenzout/go-log/fields"
	field_severity "github.com/steenzout/go-log/fields/severity"
	filter_severity "github.com/steenzout/go-log/filters/severity"
	"github.com/steenzout/go-log/formatters"
	logger_simple "github.com/steenzout/go-log/loggers/simple"
	manager_simple "github.com/steenzout/go-log/managers/simple"
)

// logger API logger manager.
var logger gol.LoggerManager

type textFormatter struct {
	formatters.Text
}

const (
	packageName = "log"
)

// Format formats the log message.
func (f textFormatter) Format(msg *gol.LogMessage) (string, error) {
	var (
		err      error
		pkg      string
		severity field_severity.Type
	)

	lmsg := msg.FieldLength()
	buffer := make([]string, lmsg, lmsg)

	i := 0
	for k, v := range *msg {
		if k != fields.Severity &&
			k != fields.Timestamp &&
			k != fields.Package {
			buffer[i] = v.(string)
			i++
		}
	}

	t, _ := msg.Timestamp()
	p, err := msg.Get(fields.Package)
	if err != nil {
		pkg = ""
	} else {
		pkg = p.(string)
	}

	message := strings.Join(buffer, " ")

	if severity, err = msg.Severity(); err != nil {
		return fmt.Sprintf("%s UNKNOWN %s %q\n", t.String(), pkg, message), nil
	}

	switch severity == field_severity.Error {
	case true:
		return fmt.Sprintf(
			"%s %s %s %s\n",
			t.String(), color.RedString("%s", severity), pkg, message), nil
	default:
		switch severity == field_severity.Info {
		case true:
			return fmt.Sprintf(
				"%s %s %s %s\n",
				t.String(), color.GreenString("%s", severity), pkg, message), nil
		default:
			return fmt.Sprintf(
				"%s %s %s %s\n",
				t.String(), severity, pkg, message), nil
		}
	}
}

var _ gol.LogFormatter = (*textFormatter)(nil)

func init() {
	logger = manager_simple.New()

	f := filter_severity.New(field_severity.Debug)
	formatter := new(textFormatter)
	log := logger_simple.New(f, formatter, os.Stdout)
	logger.Register("main", log)

	channel := make(chan *gol.LogMessage, 10)
	logger.Run(channel)
	Info(packageName, "main logger has been configured")
}

// Debug log a debug level message.
func Debug(pkg, msg string) {
	logger.Send(gol.NewDebug(
		fields.Package, pkg,
		fields.Message, msg,
	))
}

// Debugf log a formatted debug level message.
func Debugf(pkg, msg string, args ...interface{}) {
	logger.Send(gol.NewDebug(
		fields.Package, pkg,
		fields.Message, fmt.Sprintf(msg, args...)))
}

// Error log a error level message.
func Error(pkg, msg string) {
	logger.Send(gol.NewError(
		fields.Package, pkg,
		fields.Message, msg,
	))
}

// Errorf log a formatted error level message.
func Errorf(pkg, msg string, args ...interface{}) {
	logger.Send(gol.NewError(
		fields.Package, pkg,
		fields.Message, fmt.Sprintf(msg, args...)))
}

// Info log a info level message.
func Info(pkg, msg string) {
	logger.Send(gol.NewInfo(
		fields.Package, pkg,
		fields.Message, msg,
	))
}

// Infof log a formatted info level message.
func Infof(pkg, msg string, args ...interface{}) {
	logger.Send(gol.NewInfo(
		fields.Package, pkg,
		fields.Message, fmt.Sprintf(msg, args...)))
}
