package logger

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/bialas1993/etherload/pkg/cue"
	"github.com/r3labs/sse"
)

const LogFileName = "log.csv"

var (
	buffer bytes.Buffer
	f      *os.File
	mu     sync.Mutex
)

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

func Write(l *Log) {
	var d cue.Event
	json.Unmarshal(l.Event.Data, &d)

	mu.Lock()
	defer mu.Unlock()

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

	f.Write(buffer.Bytes())
}

func (l *logger) Close() {
	l.file.Close()
}
