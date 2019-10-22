package logger

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/bialas1993/cain/pkg/cue"
	"github.com/r3labs/sse"
)

const LogFileName = "log.csv"

var mu sync.Mutex

type Log struct {
	Event       *sse.Event
	Connections int
}

type logger struct {
	file *os.File
}

func New() *logger {
	f, err := os.Create(LogFileName)
	if err != nil {
		panic("logger: Can not create log file.")
	}

	f.WriteString("id,publish,receive,connections\n")

	return &logger{
		file: f,
	}
}

func (lg *logger) Write(l *Log) {

	var d cue.Event
	var buffer bytes.Buffer

	json.Unmarshal(l.Event.Data, &d)

	buffer.Write(l.Event.ID)
	buffer.WriteString(",")

	if len(d.Entries) > 0 {
		buffer.WriteString(d.Entries[0].PublishDate)
	} else {
		buffer.WriteString("-")
	}

	buffer.WriteString(",")
	buffer.WriteString(time.Now().Format("2006-01-02T15:04:05.000Z0700"))
	buffer.WriteString(",")
	buffer.WriteString(strconv.Itoa(l.Connections))
	buffer.WriteString("\n")

	mu.Lock()
	defer mu.Unlock()
	lg.file.Write(buffer.Bytes())
}

func (l *logger) Close() {
	l.file.Close()
}
