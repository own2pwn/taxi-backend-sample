package log

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

type logMsg struct {
	ErrCause string `json:"err_cause"`
}

func TestRawError(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := New("json", "info", buf)
	r, _ := http.NewRequest("POST", "google.com", nil)
	err := errors.New("test err")
	elapsed := float64(1)
	logger.NewAPIWarnLogEntry(r, err, elapsed)
	lm := new(logMsg)
	_ = json.Unmarshal(buf.Bytes(), lm)
	assert.Equal(t, "test err", lm.ErrCause)
}

func TestWrappedError(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := New("json", "info", buf)
	r, _ := http.NewRequest("POST", "google.com", nil)
	err := errors.Wrapf(errors.New("test err"), "second layer")
	elapsed := float64(1)
	logger.NewAPIWarnLogEntry(r, err, elapsed)
	lm := new(logMsg)
	_ = json.Unmarshal(buf.Bytes(), lm)
	assert.Equal(t, "test err", lm.ErrCause)
}

func TestDoubleWrappedError(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := New("json", "info", buf)
	r, _ := http.NewRequest("POST", "google.com", nil)
	err := errors.Wrapf(errors.Wrapf(errors.New("test err"), "second layer"), "third layer")
	elapsed := float64(1)
	logger.NewAPIWarnLogEntry(r, err, elapsed)
	lm := new(logMsg)
	_ = json.Unmarshal(buf.Bytes(), lm)
	assert.Equal(t, "test err", lm.ErrCause)
}
