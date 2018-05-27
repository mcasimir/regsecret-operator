package logging

import (
	"encoding/json"
	"fmt"

	logrus "github.com/sirupsen/logrus"
)

// NewLogrusJSONFormatter returns a new logrus formatter
func NewLogrusJSONFormatter() logrus.Formatter {
	return &logrusFormatter{}
}

type logrusFormatter struct{}

func (f *logrusFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields, len(entry.Data)+3)
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}

	if m, ok := data["message"]; ok {
		data["fields."+"message"] = m
	}

	if l, ok := data["level"]; ok {
		data["fields."+"level"] = l
	}

	level := entry.Level.String()

	if level == "warning" {
		level = "warn"
	}

	if level == "panic" {
		level = "fatal"
	}

	data["message"] = entry.Message
	data["level"] = level

	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}
