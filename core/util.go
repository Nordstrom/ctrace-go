package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Lines is used for testing
func (b Buffer) Lines() []string {
	return strings.Split(b.String(), "\n")
}

// Buffer is used for testing
type Buffer struct {
	bytes.Buffer
}

// SpanModel is used for testing
type SpanModel struct {
	TraceID   string                     `json:"traceId"`
	SpanID    string                     `json:"spanId"`
	ParentID  string                     `json:"parentId"`
	Operation string                     `json:"operation"`
	Start     int64                      `json:"start"`
	Finish    int64                      `json:"finish"`
	Duration  int64                      `json:"duration"`
	Tags      map[string]interface{}     `json:"tags"`
	Logs      [](map[string]interface{}) `json:"logs"`
	Baggage   map[string]string          `json:"baggage"`
}

// Spans is used for testing
func (b Buffer) Spans() []SpanModel {
	out := []SpanModel{}
	ls := b.Lines()
	for i, l := range ls {
		if l == "" {
			continue
		}
		var s SpanModel
		if err := json.Unmarshal([]byte(l), &s); err != nil {
			fmt.Printf("buf (%d lines): %s\n", len(ls), ls)
			panic("Cannot unmarshal JSON (" + err.Error() + ") for line " + strconv.Itoa(i) + l)
		}
		out = append(out, s)
	}
	return out
}
