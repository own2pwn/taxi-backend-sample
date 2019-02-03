package httprequester

import (
	"bytes"
	"context"
	_ "crypto/sha512"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/moul/http2curl"
	"github.com/pkg/errors"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/collector"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/log"
)

var (
	// ErrStatusNotOK - получили код ошибки в ответ
	ErrStatusNotOK = errors.New("Response status is not OK")
	// ErrResponseEmpty - получили пустой ответ
	ErrResponseEmpty = errors.New("Response is empty")
	// ErrParse - не смогли распарсить ответ в структуру
	ErrParse = errors.New("Cannot parse response to holder")
	// ErrCreateRequest - ошибка при создании реквеста
	ErrCreateRequest = errors.New("Cannot create request")
	// ErrReadRequest - не смогли распарсить ответ в структуру
	ErrReadRequest = errors.New("Cannot read response body")
	// ErrDoRequest - не смогли сделать запрос
	ErrDoRequest = errors.New("Cannot do request")
	// ErrContextDeadline - запрос прерван по дедлайну
	ErrContextDeadline = errors.New("Context deadline")
)

// Dict - вспомогательная структура для создания словарей хидеров и параметров
type Dict struct {
	Key   string
	Value string
}

type mock struct {
	scheme string
	host   string
	port   string
}

// Requester - структура, облегчающая запросы к сторонним сервисам по HTTP
type Requester struct {
	client      *http.Client
	logger      *log.StructuredLogger
	mockEnabled bool
	mockParams  mock
	coll        *collector.Collector
}

// NewRequester - новый реквестер HTTP
func NewRequester(client *http.Client, logger *log.StructuredLogger, coll *collector.Collector) *Requester {
	return &Requester{
		client,
		logger,
		false,
		mock{},
		coll,
	}
}

// NewMockRequester - новый реквестер HTTP
func NewMockRequester(client *http.Client, logger *log.StructuredLogger, coll *collector.Collector, scheme, host, port string) *Requester {
	return &Requester{
		client,
		logger,
		true,
		mock{
			scheme,
			host,
			port,
		},
		coll,
	}
}

func (r *Requester) createRequest(ctx context.Context, url, method string, headers, params []Dict, data []byte) (*http.Request, error) {
	var req *http.Request
	var err error
	if data != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(data))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, errors.Wrapf(ErrCreateRequest, "Cannot make %v request: %v", method, err.Error())
	}
	req = req.WithContext(ctx)
	if r.mockEnabled {
		req.Host = r.mockParams.host
		req.URL.Host = r.mockParams.host
		if r.mockParams.port != "" {
			req.Host = req.URL.Host + ":" + r.mockParams.port
			req.URL.Host = req.URL.Host + ":" + r.mockParams.port
		}
		req.URL.Scheme = r.mockParams.scheme
		req.URL.Path = fmt.Sprintf("/%v", req.Context().Value(log.CtxKeyAPIName)) + req.URL.Path
		reqID := middleware.GetReqID(ctx)
		req.Header.Set("req_id", reqID)
	}
	req.Header.Set("user-agent", "TaksaBackend/2.0")
	if headers != nil {
		for _, header := range headers {
			req.Header.Set(header.Key, header.Value)
		}
	}
	if params != nil {
		q := req.URL.Query()
		for _, param := range params {
			q.Add(param.Key, param.Value)
		}
		req.URL.RawQuery = q.Encode()
	}
	return req, nil
}

func (r *Requester) do(req *http.Request, holder interface{}) error {
	var elapsed float64
	var regionID int
	var providerName string
	start := time.Now()
	r.logger.NewAPIReqLogEntry(req)
	if provName, ok := req.Context().Value(log.CtxKeyAPIName).(string); ok {
		providerName = provName
	}
	if regID, ok := req.Context().Value(log.CtxKeyRegionID).(int); ok {
		regionID = regID
	}
	r.coll.AddProviderRequest(providerName, regionID)
	commandForLog, _ := http2curl.GetCurlCommand(req)
	r.logger.Debug(fmt.Sprintf("requesting %v %v", req.URL, commandForLog))
	response, errDo := r.client.Do(req)
	elapsed = float64(time.Since(start).Nanoseconds()) / 1000000
	if errDo != nil {
		err := errors.Wrapf(ErrDoRequest, "Error when requesting %v: %v", req.URL, errDo.Error())
		if strings.Contains(errDo.Error(), "context deadline") || strings.Contains(errDo.Error(), "context canceled") {
			err = errors.Wrapf(ErrContextDeadline, "Error when requesting %v: %v", req.URL, errDo.Error())
			r.coll.AddRequestTimeout(providerName, regionID)
		}
		r.logger.NewAPIWarnLogEntry(req, err, elapsed)
		return err
	}
	content, errRead := ioutil.ReadAll(response.Body)
	if errRead != nil {
		err := errors.Wrapf(ErrReadRequest, "GET request: error when reading pesponse body: %v", errRead.Error())
		r.logger.NewAPIWarnLogEntry(req, err, elapsed)
		r.coll.AddProviderInvalidValueResponse(providerName, fmt.Sprintf("read_%v", req.URL.Path), regionID)
		return err
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		err := errors.Wrapf(ErrStatusNotOK, "request %v, response code %v", commandForLog, response.StatusCode)
		r.logger.NewAPIWarnLogEntry(req, err, elapsed)
		r.coll.AddProviderErrorResponse(providerName, req.URL.Path, response.StatusCode, regionID)
		return err
	}
	errParse := json.NewDecoder(bytes.NewReader(content)).Decode(&holder)
	if errParse != nil {
		err := errors.Wrapf(ErrParse, "GET request: error when parsing pesponse: %v, %v", errParse.Error(), string(content))
		r.logger.NewAPIWarnLogEntry(req, err, elapsed)
		r.coll.AddProviderInvalidValueResponse(providerName, fmt.Sprintf("parse_%v", req.URL.Path), regionID)
		return err
	}
	r.logger.NewAPIDoneLogEntry(req, elapsed)
	r.coll.AddProviderResponseTime(providerName, req.URL.Path, elapsed)
	return nil
}

// Get - делаем Get запрос
func (r *Requester) Get(ctx context.Context, url string, headers []Dict, params []Dict, holder interface{}) error {
	req, err := r.createRequest(ctx, url, http.MethodGet, headers, params, nil)
	if err != nil {
		return errors.Wrapf(ErrCreateRequest, "Cannot create GET request %v", err.Error())
	}
	return r.do(req, holder)
}

// Post - делаем Post запрос
func (r *Requester) Post(ctx context.Context, url string, headers []Dict, data []byte, holder interface{}) error {
	req, err := r.createRequest(ctx, url, http.MethodPost, headers, nil, data)
	if err != nil {
		return errors.Wrapf(ErrCreateRequest, "Cannot create POST request: %v", err.Error())
	}
	r.logger.Debugf("POST, %v, data: %v", req.URL.Host, string(data))
	return r.do(req, holder)
}

// PostForm - делаем Post запрос с формой
func (r *Requester) PostForm(ctx context.Context, url string, headers []Dict, formParams []Dict, holder interface{}) error {
	// Buffer to store our request body as bytes
	var requestBody bytes.Buffer
	// Create a multipart writer
	multiPartWriter := multipart.NewWriter(&requestBody)
	for _, formParam := range formParams {
		// Populate fields
		fieldWriter, err := multiPartWriter.CreateFormField(formParam.Key)
		if err != nil {
			return errors.Wrapf(ErrCreateRequest, "Cannot create form field when post form data: %v", err.Error())
		}
		_, err = fieldWriter.Write([]byte(formParam.Value))
		if err != nil {
			return errors.Wrapf(ErrCreateRequest, "Cannot write value to form field when post form data: %v", err.Error())
		}
	}
	// We completed adding the file and the fields, let's close the multipart writer
	// So it writes the ending boundary
	multiPartWriter.Close()
	req, err := r.createRequest(ctx, url, http.MethodPost, headers, nil, requestBody.Bytes())
	if err != nil {
		return errors.Wrapf(ErrCreateRequest, "Cannot create POST FORM request: %v", err.Error())
	}
	// We need to set the content type from the writer, it includes necessary boundary as well
	req.Header.Set("Content-Type", multiPartWriter.FormDataContentType())
	return r.do(req, holder)
}
