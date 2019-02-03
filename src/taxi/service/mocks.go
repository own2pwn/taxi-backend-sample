package service

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/nburunova/taxi-backend-sample/src/infrastructure/collector"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/httprequester"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/log"
	"github.com/nburunova/taxi-backend-sample/src/product"
	"github.com/nburunova/taxi-backend-sample/src/webapi"
)

type mockAddressService struct {
	err  error
	area string
}

func newMockAddressService(err error, area string) mockAddressService {
	return mockAddressService{
		err:  err,
		area: area,
	}
}

func (m mockAddressService) Address(ctx context.Context, lat, lon float64) (*webapi.PointInfo, error) {
	return &webapi.PointInfo{"test address", 30.1, 40}, m.err
}

func (m mockAddressService) AreaNameByLatLon(lat, lon float64) string {
	return m.area
}

type mockProductCache struct {
	prods []product.Product
	err   error
}

func newMockProductCache(err error, prods ...product.Product) ProductsCache {
	return mockProductCache{
		prods: prods,
		err:   err,
	}
}

func (m mockProductCache) GetProducts(regionID int) ([]product.Product, error) {
	return m.prods, m.err
}

func (m mockProductCache) IsOK() bool {
	return true
}

var tmpVars = map[string]string{
	"%from.lat%":     "0.0",
	"%from.lon%":     "0.1",
	"%from.address%": "address1",
	"%to.lat%":       "1.0",
	"%to.lon%":       "1.1",
	"%to.address%":   "address2",
}

type mockAPIDataGetter struct {
	apiDatas []APIData
	err      error
	name     string
}

func newMockAPIDataGetter(err error, name string, apiDatas ...APIData) APIDataGetter {
	return mockAPIDataGetter{
		apiDatas: apiDatas,
		err:      err,
		name:     name,
	}
}

func (m mockAPIDataGetter) GetAPIData(ctx context.Context, httpreq *httprequester.Requester, taxiReq Request) ([]APIData, error) {
	return m.apiDatas, m.err
}

func (m mockAPIDataGetter) APIName() string {
	return m.name
}

type mockDistanceTimeService struct {
	distance int
	time     int
	err      error
}

func newMockDistanceTimeService(err error, distance, time int) DistanceTimeService {
	return mockDistanceTimeService{
		distance: distance,
		time:     time,
		err:      err,
	}
}

func (m mockDistanceTimeService) DistanceTime(ctx context.Context, httpreq *httprequester.Requester, taxiReq Request) (int, int, error) {
	return m.distance, m.time, m.err
}

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

func getTestService() *Service {
	provName := "test1"
	prod := product.Product{}
	prod.ProviderName = provName
	pcache := newMockProductCache(nil, prod)

	var testAPIData1 = APIData{}
	dg := []APIDataGetter{
		newMockAPIDataGetter(nil, provName, testAPIData1),
	}
	dt := newMockDistanceTimeService(nil, 1000, 2000)
	addr := newMockAddressService(nil, "")
	return NewBuilder().
		WithAPIs(dg).
		WithProductCache(pcache).
		WithRequester(testHTTPRequester).
		WithDistanceTimeSrv(dt).
		WithAddressSrv(addr).
		WithStatCollector(testCollector).
		WithLogger(testLogger).
		Build()
}

var testTaxiRequestMoscow = Request{
	RegionID: 32,
	Point1: Point{
		Lon:     37.610621,
		LonStr:  "37.610621",
		Lat:     55.750376,
		LatStr:  "55.750376",
		Address: "Тестовая точка 1",
	},
	Point2: Point{
		Lon:     37.62002,
		LonStr:  "37.62002",
		Lat:     55.760736,
		LatStr:  "55.760736",
		Address: "Тестовая точка 2",
	},
	OnlyAPI: true,
}

func strPointer(value string) *string {
	s := value
	return &s
}

func urlPointer(value string) *url.URL {
	u, _ := url.Parse(value)
	return u
}

func float32Pointer(value float32) *float32 {
	f := value
	return &f
}

func intPointer(value int) *int {
	i := value
	return &i
}
