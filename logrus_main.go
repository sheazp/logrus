package xlogrus

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/pkg/errors"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

// config logrus log to local filesystem, with file rotation
func InitLogger(logPath string, logFileName string, maxAge time.Duration, rotationTime time.Duration) bool {
	baseLogPaht := path.Join(logPath, logFileName)
	infoWriter, err := rotatelogs.New(
		baseLogPaht+".info.%Y-%m-%d_%H_%M_%S",
		rotatelogs.WithMaxAge(maxAge),             // 文件最大保存时间
		rotatelogs.WithRotationTime(rotationTime), // 日志切割时间间隔
	)
	errorWriter, err := rotatelogs.New(
		baseLogPaht+".error.%Y-%m-%d_%H_%M_%S",
		rotatelogs.WithMaxAge(maxAge),             // 文件最大保存时间
		rotatelogs.WithRotationTime(rotationTime), // 日志切割时间间隔
	)
	if err != nil {
		log.Errorf("config local file system logger error. %+v", errors.WithStack(err))
	}

	formatter := &MyFormatter{Prefix: "Gwell"}
	lfHook := lfshook.NewHook(lfshook.WriterMap{
		log.DebugLevel: infoWriter, // 为不同级别设置不同的输出目的
		log.InfoLevel:  infoWriter,
		log.WarnLevel:  infoWriter,
		log.ErrorLevel: errorWriter,
		log.FatalLevel: errorWriter,
		log.PanicLevel: errorWriter,
	}, formatter)
	log.AddHook(lfHook)
	log.SetReportCaller(true)
	log.SetFormatter(formatter)
	return true
}

func main() {
	InitLogger("./log", "test.log", 5*time.Minute, 3*time.Minute)
	for {
		log.Debugf("mylog %v\n", time.Now().Format("2006/01/02 15:04:05.999"))
		log.Printf("mylog %v\n", time.Now().Format("2006/01/02 15:04:05.999"))
		log.Warnf("mylog %v\n", time.Now().Format("2006/01/02 15:04:05.999"))
		log.Errorf("mylog %v\n", time.Now().Format("2006/01/02 15:04:05.999"))
		//log.Fatal("mylog %v\n", time.Now().Format("2006/01/02 15:04:05.999"))
		time.Sleep(2 * time.Second)
	}
	log.Info("Exit\n")
}

// MyFormatter 自定义 formatter
type MyFormatter struct {
	Prefix string
}

// Format implement the Formatter interface
func (mf *MyFormatter) Format(entry *log.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	ostr := fmt.Sprintf("%s [%v] ", time.Now().Format("2006/01/02 15:04:05.999"), entry.Level.String())
	if lv, err := log.ParseLevel(entry.Level.String()); err == nil && lv <= log.ErrorLevel {
		//log大于等于error时，需打印文件名和行号
		if entry.Caller != nil {
			i := strings.LastIndexAny(entry.Caller.File, "/")
			if i >= 0 {
				i++
			} else {
				i = 0
			}
			ostr += fmt.Sprintf("[%v:%d]", entry.Caller.File[i:], entry.Caller.Line)
		}
	}
	b.WriteString(ostr + entry.Message) // entry.Message 就是需要打印的日志

	return b.Bytes(), nil
}
