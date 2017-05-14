package events

import (
	"encoding/json"
	"os"

	"github.com/alphasoc/namescore/client"
)

// Writer interface for log api events response.
type Writer interface {
	Write(*client.EventsResponse) error
}

// JSONFileWriter implements Writer interface and write
// api events in JSON format.
type JSONFileWriter struct {
	f *os.File
}

// NewJSONFileWriter creates new json file logger.
func NewJSONFileWriter(file string) (*JSONFileWriter, error) {
	switch file {
	case "stdout":
		return &JSONFileWriter{os.Stdout}, nil
	case "stderr":
		return &JSONFileWriter{os.Stderr}, nil
	default:
		f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
		return &JSONFileWriter{f}, nil
	}
}

// Write writes event response to file in json format
func (l *JSONFileWriter) Write(e *client.EventsResponse) error {
	// do not log if there is no events
	if len(e.Events) == 0 {
		return nil
	}

	b, err := json.Marshal(e)
	if err != nil {
		return err
	}

	if _, err = l.f.Write(b); err != nil {
		return err
	}
	_, err = l.f.Write([]byte("\n"))
	return err
}

// Close closes the File. It returns an error, if any.
func (l *JSONFileWriter) Close() error {
	return l.f.Close()
}
