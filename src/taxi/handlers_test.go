package taxi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"context"

	"github.com/pkg/errors"
	"github.com/nburunova/taxi-backend-sample/src/taxi/service"
)

type mockServiceResponseError struct{}

func (m mockServiceResponseError) ParseTaxiRequest(ctx context.Context, r *http.Request) (service.Request, error) {
	return service.Request{}, nil
}
func (m mockServiceResponseError) Response(ctx context.Context, req service.Request, priceCoeff float64) (*service.Response, error) {
	return &service.Response{}, errors.New("Mocked Service Response Fail")
}

func (m mockServiceResponseError) EvaluateTaxiRequest(ctx context.Context, taxiReq *service.Request) error {
	return nil
}

type mockServiceTaxiReqError struct{}

func (m mockServiceTaxiReqError) ParseTaxiRequest(ctx context.Context, r *http.Request) (service.Request, error) {
	return service.Request{}, errors.New("Mocked Service Taxi request parsing fail")
}
func (m mockServiceTaxiReqError) Response(ctx context.Context, req service.Request, priceCoeff float64) (*service.Response, error) {
	return &service.Response{}, nil
}
func (m mockServiceTaxiReqError) EvaluateTaxiRequest(ctx context.Context, taxiReq *service.Request) error {
	return nil
}

func TestResponseError(t *testing.T) {
	s := mockServiceResponseError{}
	testHandler := SomeHandler(s, 1.3, map[int]float64{99: 1.0}, 4, Handler)

	payload := strings.NewReader("{\r\n    \"region_id\": 14, \r\n    \"point1\": { \r\n        \"lon\": 30.723449,\r\n        \"lat\": 46.441982\r\n    },\r\n    \"point2\": { \r\n        \"lon\": 30.759315,\r\n        \"lat\": 46.453352\r\n    },\r\n    \"only_api\": true \r\n}")
	req, err := http.NewRequest("POST", "/taksa/api/1.0/route/calculate", payload)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
}

func TestRequestError(t *testing.T) {
	s := mockServiceResponseError{}
	testHandler := SomeHandler(s, 1.3, map[int]float64{99: 1.0}, 4, Handler)

	payload := strings.NewReader("{\r\n    \"region_id\": 14, \r\n    \"point1\": { \r\n        \"lon\": 30.723449,\r\n        \"lat\": 46.441982\r\n    },\r\n    \"point2\": { \r\n        \"lon\": 30.759315,\r\n        \"lat\": 46.453352\r\n    },\r\n    \"only_api\": true \r\n}")
	req, err := http.NewRequest("POST", "/taksa/api/1.0/route/calculate", payload)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
}
