package logger

import (
	"testing"

	"github.com/r3labs/sse"
)

func BenchmarkLogger(b *testing.B) {
	l := New()

	for n := 0; n < b.N; n++ {
		l.Write(&Log{
			Event:       &sse.Event{ID: []byte("1")},
			Connections: 100})
	}
}
