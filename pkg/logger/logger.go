package logger

import (
	"os"
	"sync"
	"github.com/bialas1993/etherload/pkg/cue"
	"github.com/r3labs/sse"
)

type logger struct {
	file *os.File
}

const LogFileName = "log.csv"

var (
	buffer bytes.Buffer
	f *os.File
	mu sync.Mutex
)

func New() *logger {
	f, err := os.Create(LogFileName)
	if err != nil {
		panic("Can not create log file.")
	}

	return &logger{
		file: f,
	}
}

func Write(event *sse.Event) {
	var d cue.Event
	json.Unmarshal(e.Data, &d)

	mu.Lock()
	defer mu.Unlock()

	buffer.Write(e.ID)
	buffer.WriteString(",")

	if len(d.Entries) > 0 {
		buffer.WriteString(d.Entries[0].PublishDate)
	} else {
		buffer.WriteString("-")
	}

	buffer.WriteString(",")
	buffer.WriteString(time.Now().Format("2006-01-02T15:04:05.000Z0700"))
	buffer.WriteString(",")
	buffer.WriteString(strconv.Itoa(openedConnections))
	buffer.WriteString("\n")

	f.Write(buffer.Bytes())
}

func (l *logger) Close() {
	l.file.Close()
}
