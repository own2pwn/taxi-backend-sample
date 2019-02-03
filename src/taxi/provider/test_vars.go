package provider

import (
	"context"
	"net/http"
	"time"

	"github.com/nburunova/taxi-backend-sample/src/infrastructure/collector"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/httprequester"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/log"
	"github.com/nburunova/taxi-backend-sample/src/taxi/service"
)

func newTestContext() context.Context {
	testContext := context.Background()
	testContext = context.WithValue(testContext, log.CtxKeyAPIName, "test")
	testContext = context.WithValue(testContext, log.CtxKeyRegionID, 1)
	return testContext
}

var testContext = newTestContext()

var testHttpClient = &http.Client{
	Transport: &http.Transport{
		// Максимальное время бездействия до закрытия соединения; сколько времи простаивающее соединение хранится в пуле
		IdleConnTimeout: 5 * time.Second,
	},
}

var testLogger = log.NewEmpty()

var testCollector = collector.NewCollector(testLogger)

var testHTTPRequester = httprequester.NewRequester(testHttpClient, testLogger, testCollector)

var testTaxiRequestMoscow = service.Request{
	ReqID:    "123",
	RegionID: 32,
	Point1: service.Point{
		Lon:     37.610621,
		LonStr:  "37.610621",
		Lat:     55.750376,
		LatStr:  "55.750376",
		Address: "Тестовая точка 1",
	},
	Point2: service.Point{
		Lon:     37.62002,
		LonStr:  "37.62002",
		Lat:     55.760736,
		LatStr:  "55.760736",
		Address: "Тестовая точка 2",
	},
	OnlyAPI: true,
}
