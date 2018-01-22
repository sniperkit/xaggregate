package github

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/sniperkit/xstats/pkg"
	"github.com/sniperkit/xtask/plugin/counter"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var (
	log                  = logrus.New()
	counters *counter.Oc = counter.NewOc()
)

/*
// Alias these functions. More should be added here as needed.
var (
	Error   = logrus.Error
	Errorf  = logrus.Errorf
	Errorln = logrus.Errorln
	Fatal   = logrus.Fatal
	Fatalf  = logrus.Fatalf
	Fatalln = logrus.Fatalln
	Panic   = logrus.Panic
	Panicf  = logrus.Panicf
	Panicln = logrus.Panicln
	Print   = logrus.Print
	Printf  = logrus.Printf
	Println = logrus.Println
	Debug   = logrus.Debug
	Debugf  = logrus.Debugf
	Debugln = logrus.Debugln
	Warn    = logrus.Warn
	Warnf   = logrus.Warnf
	Warnln  = logrus.Warnln
	Info    = logrus.Info
	Infof   = logrus.Infof
	Infoln  = logrus.Infoln
)
*/

// type logFields logrus.Fields
type Fields logrus.Fields

// WithFields is an alias for logrus.WithFields.
func LogWithFields(f Fields) *logrus.Entry {
	return logrus.WithFields(logrus.Fields(f))
}

type funcMetrics struct {
	calls struct {
		count  int           `metric:"count" type:"counter"`
		failed int           `metric:"failed" type:"counter"`
		time   time.Duration `metric:"time"  type:"histogram"`
	} `metric:"func.calls"`
}

func GetCaller() string {
	_, file, line, _ := runtime.Caller(3)
	return fmt.Sprintf("%s:%d", trimPath(file), line)
}

func trimPath(path string) string {
	// For details, see https://github.com/uber-go/zap/blob/e15639dab1b6ca5a651fe7ebfd8d682683b7d6a8/zapcore/entry.go#L101
	if idx := strings.LastIndexByte(path, '/'); idx >= 0 {
		if idx := strings.LastIndexByte(path[:idx], '/'); idx >= 0 {
			// Keep everything after the penultimate separator.
			return path[idx+1:]
		}
	}
	return path
}

func addMetrics(start time.Time, incr int, failed bool) {
	callTime := time.Now().Sub(start)
	m := &funcMetrics{}
	m.calls.count = incr
	m.calls.time = callTime
	if failed {
		m.calls.failed = incr
	}
	stats.Report(m)
}

// return the source filename after the last slash
func chopPath(original string) string {
	i := strings.LastIndex(original, "/")
	if i == -1 {
		return original
	} else {
		return original[i+1:]
	}
}

func funcTrack(start time.Time) {
	return
	function, file, line, _ := runtime.Caller(1)
	// go func() {
	elapsed := time.Since(start)
	log.Printf("main().funcTrack() %s took %s", fmt.Sprintf("%s:%s:%d", runtime.FuncForPC(function).Name(), chopPath(file), line), elapsed)
	// addMetrics(start, 1, err != nil)
	//}()
}

func counterTrack(name string, incr int) {
	go func() {
		counters.Increment(name, incr)
	}()
}

func init() {
	log.Formatter = new(prefixed.TextFormatter)
	log.Level = logrus.DebugLevel
}
