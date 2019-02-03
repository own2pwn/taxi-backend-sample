package log

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func tsNow() string {
	return time.Now().UTC().Format(time.StampMilli)
}

// CtxKey - тип ключа контекста
type CtxKey string

// CtxKeyAPIName - ключ контекста для имени апи
var CtxKeyAPIName = CtxKey("APIName")

// CtxKeyRegionID - ключ контекста для имени региона
var CtxKeyRegionID = CtxKey("RegionID")

// Logger - интерфейс для логгера
type Logger interface {
	logrus.FieldLogger
}

// NewEmpty - возвращает логгер, который не логгирует(используется в тестах)
func NewEmpty() *StructuredLogger {
	logger := logrus.New()
	logger.Out = ioutil.Discard
	return &StructuredLogger{logger}
}

// New - создаем новый логгер
func New(format, level string, output io.Writer) *StructuredLogger {
	logger := logrus.New()

	logger.Out = output

	f := strings.ToLower(format)
	switch f {
	case "json":
		logger.Formatter = &logrus.JSONFormatter{}
	case "text":
		logger.Formatter = &logrus.TextFormatter{ForceColors: true}
	default:
		logger.Warnf("log: invalid formatter: %v, continue with default", f)
	}

	l := strings.ToLower(level)
	sev, err := logrus.ParseLevel(l)
	if err != nil {
		logger.Warnf("log: invalid level: %v, continue with info", l)
		sev = logrus.InfoLevel
	}
	logger.Level = sev

	return &StructuredLogger{logger}
}

// StructuredLogger - структура логгера с расширенными полями и методами
type StructuredLogger struct {
	*logrus.Logger
}

// NewLogEntry - логируем запрос к сервису такси
func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &StructuredLoggerEntry{Logger: logrus.NewEntry(l.Logger)}
	entry.Logger = entry.Logger.WithFields(logrus.Fields{
		"message_type": "request_incoming",
		"ts":           tsNow(),
		"req_id":       middleware.GetReqID(r.Context()),
		"http_method":  r.Method,
		"remote_addr":  r.RemoteAddr,
		"url":          r.URL,
		"user_agent":   r.UserAgent(),
		"uri":          r.RequestURI,
	})
	entry.Logger.Infoln("request started")
	return entry
}

// NewAPIReqLogEntry - логируем запрос к провайдерам такси
func (l *StructuredLogger) NewAPIReqLogEntry(r *http.Request) {
	aName := r.Context().Value(CtxKeyAPIName)
	l.WithFields(logrus.Fields{
		"message_type": fmt.Sprintf("request_%v", aName),
		"ts":           tsNow(),
		"req_id":       middleware.GetReqID(r.Context()),
		"http_method":  r.Method,
		"url":          r.URL,
		"uri":          r.RequestURI,
	}).Infoln("API request started")
}

// NewAPIWarnLogEntry - логируем ошибку к провайдерам такси
func (l *StructuredLogger) NewAPIWarnLogEntry(r *http.Request, err error, elapsed float64) {
	aName := r.Context().Value(CtxKeyAPIName)
	errCause := errors.Cause(err)
	l.WithFields(logrus.Fields{
		"message_type": fmt.Sprintf("error_%v", aName),
		"err_cause":    errCause,
		"ts":           tsNow(),
		"req_id":       middleware.GetReqID(r.Context()),
		"http_method":  r.Method,
		"url":          r.URL,
		"uri":          r.RequestURI,
		"elapsed":      elapsed,
	}).Warningf("API request failed: %v\n", err.Error())
}

// NewAPIDoneLogEntry - логируем окончание запроса к провайдеру такси
func (l *StructuredLogger) NewAPIDoneLogEntry(r *http.Request, elapsed float64) {
	aName := r.Context().Value(CtxKeyAPIName)
	l.WithFields(logrus.Fields{
		"message_type": fmt.Sprintf("response_%v", aName),
		"ts":           tsNow(),
		"req_id":       middleware.GetReqID(r.Context()),
		"http_method":  r.Method,
		"url":          r.URL,
		"uri":          r.RequestURI,
		"elapsed":      elapsed,
	}).Infoln("API request done")
}

//TimingLogEntry - засекаем время
func (l *StructuredLogger) TimingLogEntry(ctx context.Context, elapsed float64, message string) {
	l.WithFields(logrus.Fields{
		"message_type": "timings",
		"req_id":       middleware.GetReqID(ctx),
		"elapsed":      elapsed,
	}).Debugln(message)
}

//ServiceWarningLogEntry - записываем ошибку сервиса в лог
func (l *StructuredLogger) ServiceWarningLogEntry(ctx context.Context, err error, message string, msgType string) {
	errCause := errors.Cause(err)
	l.WithFields(logrus.Fields{
		"message_type": fmt.Sprintf("service_%v", msgType),
		"err_cause":    errCause,
		"req_id":       middleware.GetReqID(ctx),
		"message":      message,
	}).Warningln(err.Error())
}

// StructuredLoggerEntry - струтура описывающая запись в логе
type StructuredLoggerEntry struct {
	Logger logrus.FieldLogger
}

func (l *StructuredLoggerEntry) Write(status, bytes int, elapsed time.Duration) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"resp_status": status, "resp_bytes_length": bytes,
		"resp_elapsed_ms": float64(elapsed.Nanoseconds()) / 1000000.0,
		"message_type":    "request_complete",
		"ts":              tsNow(),
	})

	l.Logger.Infoln("request complete")
}

// Panic - логируем панику
func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"stack": string(stack),
		"panic": fmt.Sprintf("%+v", v),
		"ts":    tsNow(),
	})
}
