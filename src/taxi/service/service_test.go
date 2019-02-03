package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/nburunova/taxi-backend-sample/src/product"
)

var priceOff = 1.3

func TestResponseFields(t *testing.T) {
	avgEta := 10
	provName := "test1"
	prod := product.Product{
		ID:             1,
		RegionID:       1,
		Name:           "test",
		Tariffs:        nil,
		Title:          "test gett",
		ShortTitle:     strPointer("t gett"),
		SiteCaption:    strPointer("test gett site"),
		SiteValue:      strPointer("www.test.com"),
		AppURLTemplate: strPointer("careem://?from_lat=%from.lat%,from_lon=%from.lon%,from_addr=%from.address%.to_lat=%to.lat%,to_lon=%to.lon%,to_addr=%to.address%"),
		PhoneCaption:   strPointer("cap123456789"),
		PhoneValue:     strPointer("123456789"),
		AndroidAppURL:  strPointer("http://android.app"),
		AndroidAppID:   strPointer("android app id"),
		IosAppURL:      strPointer("http://ios.app"),
		IosAppID:       strPointer("ios app id"),
		APIOrgID:       100200300,
		APIID:          400500600,
		APIData:        strPointer("api data"),
		Rating:         float32Pointer(5.6),
		AvgEta:         intPointer(10),
		ProviderName:   "test",
		CurrencyCode:   strPointer("RUB"),
		ImageURL:       urlPointer("http://my_image.com"),
	}
	prod.AvgEta = &avgEta
	prod.ProviderName = provName
	pcache := newMockProductCache(nil, prod)
	var tmpVars = map[string]string{
		"%from.lat%":     "0.0",
		"%from.lon%":     "0.1",
		"%from.address%": "адрес1",
		"%to.lat%":       "1.0",
		"%to.lon%":       "1.1",
		"%to.address%":   "address2",
	}
	var testAPIData1 = APIData{
		ProductID:    "1",
		TariffName:   "tariff1",
		DisplayName:  "displayName1",
		PriceMin:     100,
		PriceMax:     300,
		PriceMean:    200.0,
		Eta:          5,
		TemplateVars: tmpVars,
	}
	dg := []APIDataGetter{
		newMockAPIDataGetter(nil, provName, testAPIData1),
	}
	dt := newMockDistanceTimeService(nil, 1000, 2000)
	addr := newMockAddressService(nil, "")

	service := NewBuilder().
		WithAPIs(dg).
		WithProductCache(pcache).
		WithRequester(testHTTPRequester).
		WithDistanceTimeSrv(dt).
		WithAddressSrv(addr).
		WithStatCollector(testCollector).
		WithLogger(testLogger).
		Build()
	resp, err := service.Response(testContext, testTaxiRequestMoscow, priceOff)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(resp.Result.Optimal.Results))
	//Required fields
	assert.Equal(t, 200, resp.Result.Optimal.Results[0].Price)
	assert.Equal(t, 300, *resp.Result.Optimal.Results[0].PriceRanges.Max)
	assert.Equal(t, 100, *resp.Result.Optimal.Results[0].PriceRanges.Min)
	assert.Equal(t, 5, *resp.Result.Optimal.Results[0].Eta)
	assert.Equal(t, avgEta, *resp.Result.Optimal.Results[0].AvgEta)
	assert.Equal(t, "RUB", *resp.Result.Optimal.Results[0].CurrencyCode)
	assert.Equal(t, 1000, *resp.Meta.Distance)
	assert.Equal(t, 2000, *resp.Meta.Time)
	//Operator Required Fields
	assert.Equal(t, "careem://?from_lat=0.0%2Cfrom_lon%3D0.1%2Cfrom_addr%3D%D0%B0%D0%B4%D1%80%D0%B5%D1%811.to_lat%3D1.0%2Cto_lon%3D1.1%2Cto_addr%3Daddress2", *resp.Result.Optimal.Results[0].Operator.URL)
	assert.Equal(t, "www.test.com", *resp.Result.Optimal.Results[0].Operator.Site.Value)
	assert.Equal(t, "http://ios.app", resp.Result.Optimal.Results[0].Operator.StoreURLs.Ios.URL)
	assert.Equal(t, "http://android.app", resp.Result.Optimal.Results[0].Operator.StoreURLs.Android.URL)
	assert.Equal(t, "100200300", *resp.Result.Optimal.Results[0].Operator.OrgID)
	assert.Equal(t, "400500600", *resp.Result.Optimal.Results[0].Operator.BranchID)
	assert.Equal(t, "t gett", *resp.Result.Optimal.Results[0].Operator.ShortTitle)
	assert.Equal(t, "displayName1", *resp.Result.Optimal.Results[0].Operator.Title)
	assert.Equal(t, "http://my_image.com", *resp.Result.Optimal.Results[0].Operator.Image)
}

func TestNoOptimalAndNoOptimalResponses(t *testing.T) {
	provName := "test1"
	prod := product.Product{
		ProviderName: provName,
		ID:           1,
		RegionID:     1,
		Name:         "test",
		IsOptimal:    false,
	}
	pcache := newMockProductCache(nil, prod)
	var testAPIData1 = APIData{
		PriceMean: 200.0,
	}
	dg := []APIDataGetter{
		newMockAPIDataGetter(nil, provName, testAPIData1),
	}
	dt := newMockDistanceTimeService(nil, 1000, 2000)
	addr := newMockAddressService(nil, "")

	service := NewBuilder().
		WithAPIs(dg).
		WithProductCache(pcache).
		WithRequester(testHTTPRequester).
		WithDistanceTimeSrv(dt).
		WithAddressSrv(addr).
		WithStatCollector(testCollector).
		WithLogger(testLogger).
		Build()
	resp, err := service.Response(testContext, testTaxiRequestMoscow, priceOff)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(resp.Result.Optimal.Results))
}

func TestOptimalAndOptimalResponses(t *testing.T) {
	provName := "test1"
	prod := product.Product{
		ProviderName: provName,
		ID:           1,
		RegionID:     1,
		Name:         "test",
		IsOptimal:    true,
	}
	pcache := newMockProductCache(nil, prod)
	var testAPIData1 = APIData{
		PriceMean: 200.0,
	}
	dg := []APIDataGetter{
		newMockAPIDataGetter(nil, provName, testAPIData1),
	}
	dt := newMockDistanceTimeService(nil, 1000, 2000)
	addr := newMockAddressService(nil, "")

	service := NewBuilder().
		WithAPIs(dg).
		WithProductCache(pcache).
		WithRequester(testHTTPRequester).
		WithDistanceTimeSrv(dt).
		WithAddressSrv(addr).
		WithStatCollector(testCollector).
		WithLogger(testLogger).
		Build()
	resp, err := service.Response(testContext, testTaxiRequestMoscow, priceOff)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(resp.Result.Optimal.Results))
}

func TestOptimalAndNoOptimalResponses(t *testing.T) {
	provName1 := "test1"
	provName2 := "test2"
	prod1 := product.Product{
		ProviderName: provName1,
		ID:           1,
		RegionID:     1,
		Name:         "test1",
		IsOptimal:    true,
	}
	prod2 := product.Product{
		ProviderName: provName2,
		ID:           2,
		RegionID:     1,
		Name:         "test2",
		IsOptimal:    false,
	}
	pcache := newMockProductCache(nil, prod1, prod2)
	var testAPIData1 = APIData{
		PriceMean: 200.0,
	}
	dg := []APIDataGetter{
		newMockAPIDataGetter(nil, provName2, testAPIData1),
	}
	dt := newMockDistanceTimeService(nil, 1000, 2000)
	addr := newMockAddressService(nil, "")

	service := NewBuilder().
		WithAPIs(dg).
		WithProductCache(pcache).
		WithRequester(testHTTPRequester).
		WithDistanceTimeSrv(dt).
		WithAddressSrv(addr).
		WithStatCollector(testCollector).
		WithLogger(testLogger).
		Build()
	resp, err := service.Response(testContext, testTaxiRequestMoscow, priceOff)
	assert.Nil(t, err)
	assert.Nil(t, resp.Result.Optimal)
	assert.Equal(t, len(resp.Result.Else.Results), 1)
}

func TestNoMatchTariff(t *testing.T) {
	provName := "test1"
	prod := product.Product{}
	prod.Tariffs = []string{"tar1"}
	prod.ProviderName = provName
	pcache := newMockProductCache(nil, prod)

	var testAPIData1 = APIData{}
	var testAPIData2 = APIData{}
	var testAPIData3 = APIData{}
	dg := []APIDataGetter{
		newMockAPIDataGetter(nil, provName, testAPIData1, testAPIData2, testAPIData3),
	}
	dt := newMockDistanceTimeService(nil, 1000, 2000)
	addr := newMockAddressService(nil, "")

	service := NewBuilder().
		WithAPIs(dg).
		WithProductCache(pcache).
		WithRequester(testHTTPRequester).
		WithDistanceTimeSrv(dt).
		WithAddressSrv(addr).
		WithStatCollector(testCollector).
		WithLogger(testLogger).
		Build()
	resp, err := service.Response(testContext, testTaxiRequestMoscow, priceOff)
	assert.NotNil(t, err)
	assert.Nil(t, resp.Result.Optimal)
	assert.Nil(t, resp.Result.Else)
}

func TestEmtyTariff(t *testing.T) {
	provName := "test1"
	prod := product.Product{}
	prod.Tariffs = nil
	prod.ProviderName = provName
	pcache := newMockProductCache(nil, prod)

	var testAPIData1 = APIData{}
	var testAPIData2 = APIData{}
	var testAPIData3 = APIData{}
	dg := []APIDataGetter{
		newMockAPIDataGetter(nil, provName, testAPIData1, testAPIData2, testAPIData3),
	}
	dt := newMockDistanceTimeService(nil, 1000, 2000)
	addr := newMockAddressService(nil, "")

	service := NewBuilder().
		WithAPIs(dg).
		WithProductCache(pcache).
		WithRequester(testHTTPRequester).
		WithDistanceTimeSrv(dt).
		WithAddressSrv(addr).
		WithStatCollector(testCollector).
		WithLogger(testLogger).
		Build()
	resp, err := service.Response(testContext, testTaxiRequestMoscow, priceOff)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(resp.Result.Optimal.Results))
	assert.Equal(t, 2, len(resp.Result.Else.Results))
}

func TestManyTariffs(t *testing.T) {
	provName := "test1"
	prod := product.Product{}
	prod.Tariffs = []string{"tariff1", "tariff2"}
	prod.ProviderName = provName
	pcache := newMockProductCache(nil, prod)

	var testAPIData1 = APIData{}
	testAPIData1.TariffName = "tariff1"
	var testAPIData2 = APIData{}
	testAPIData2.TariffName = "tariff2"
	var testAPIData3 = APIData{}
	testAPIData3.TariffName = "tariff3"
	dg := []APIDataGetter{
		newMockAPIDataGetter(nil, provName, testAPIData1, testAPIData2, testAPIData3),
	}
	dt := newMockDistanceTimeService(nil, 1000, 2000)
	addr := newMockAddressService(nil, "")

	service := NewBuilder().
		WithAPIs(dg).
		WithProductCache(pcache).
		WithRequester(testHTTPRequester).
		WithDistanceTimeSrv(dt).
		WithAddressSrv(addr).
		WithStatCollector(testCollector).
		WithLogger(testLogger).
		Build()
	resp, err := service.Response(testContext, testTaxiRequestMoscow, priceOff)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(resp.Result.Optimal.Results))
	assert.Equal(t, 1, len(resp.Result.Else.Results))
}

func TestSort(t *testing.T) {
	provName := "test1"
	prod := product.Product{}
	prod.ProviderName = provName
	pcache := newMockProductCache(nil, prod)

	var testAPIData1 = APIData{}
	testAPIData1.PriceMean = 500
	var testAPIData2 = APIData{}
	testAPIData2.PriceMean = 100
	var testAPIData3 = APIData{}
	testAPIData3.PriceMean = 200
	dg := []APIDataGetter{
		newMockAPIDataGetter(nil, provName, testAPIData1, testAPIData2, testAPIData3),
	}
	dt := newMockDistanceTimeService(nil, 1000, 2000)
	addr := newMockAddressService(nil, "")

	service := NewBuilder().
		WithAPIs(dg).
		WithProductCache(pcache).
		WithRequester(testHTTPRequester).
		WithDistanceTimeSrv(dt).
		WithAddressSrv(addr).
		WithStatCollector(testCollector).
		WithLogger(testLogger).
		Build()
	resp, err := service.Response(testContext, testTaxiRequestMoscow, priceOff)
	assert.Nil(t, err)
	assert.Equal(t, 100, resp.Result.Optimal.Results[0].Price)
	assert.Equal(t, 200, resp.Result.Else.Results[0].Price)
	assert.Equal(t, 500, resp.Result.Else.Results[1].Price)
}

func TestSortCoeff(t *testing.T) {
	provName := "test1"
	prod := product.Product{}
	prod.ProviderName = provName
	pcache := newMockProductCache(nil, prod)

	minVal := 90.0
	maxVal := minVal * priceOff

	var testAPIData1 = APIData{}
	testAPIData1.PriceMean = 500
	var testAPIData2 = APIData{}
	testAPIData2.PriceMean = minVal
	var testAPIData3 = APIData{}
	testAPIData3.PriceMean = maxVal
	dg := []APIDataGetter{
		newMockAPIDataGetter(nil, provName, testAPIData1, testAPIData2, testAPIData3),
	}
	dt := newMockDistanceTimeService(nil, 1000, 2000)
	addr := newMockAddressService(nil, "")

	service := NewBuilder().
		WithAPIs(dg).
		WithProductCache(pcache).
		WithRequester(testHTTPRequester).
		WithDistanceTimeSrv(dt).
		WithAddressSrv(addr).
		WithStatCollector(testCollector).
		WithLogger(testLogger).
		Build()
	resp, err := service.Response(testContext, testTaxiRequestMoscow, priceOff)
	epsilon := priceOff - 1
	assert.Nil(t, err)
	assert.InEpsilon(t, minVal, resp.Result.Optimal.Results[0].Price, epsilon)
	assert.InEpsilon(t, minVal, resp.Result.Else.Results[0].Price, epsilon)
	assert.Equal(t, 500, resp.Result.Else.Results[1].Price)
}

func TestProductCacheErr(t *testing.T) {
	provName := "test1"
	prod := product.Product{}
	prod.ProviderName = provName
	prodCacheErr := errors.New("Product Cache err")
	pcache := newMockProductCache(prodCacheErr, prod)

	var testAPIData1 = APIData{}
	dg := []APIDataGetter{
		newMockAPIDataGetter(nil, provName, testAPIData1),
	}
	dt := newMockDistanceTimeService(nil, 1000, 2000)
	addr := newMockAddressService(nil, "")

	service := NewBuilder().
		WithAPIs(dg).
		WithProductCache(pcache).
		WithRequester(testHTTPRequester).
		WithDistanceTimeSrv(dt).
		WithAddressSrv(addr).
		WithStatCollector(testCollector).
		WithLogger(testLogger).
		Build()
	_, err := service.Response(testContext, testTaxiRequestMoscow, priceOff)
	assert.NotNil(t, err)
}

func TestProductCacheEmpty(t *testing.T) {
	prod := product.Product{}
	pcache := newMockProductCache(nil, prod)

	var testAPIData1 = APIData{}
	dg := []APIDataGetter{
		newMockAPIDataGetter(nil, "test1", testAPIData1),
	}
	dt := newMockDistanceTimeService(nil, 1000, 2000)
	addr := newMockAddressService(nil, "")

	service := NewBuilder().
		WithAPIs(dg).
		WithProductCache(pcache).
		WithRequester(testHTTPRequester).
		WithDistanceTimeSrv(dt).
		WithAddressSrv(addr).
		WithStatCollector(testCollector).
		WithLogger(testLogger).
		Build()
	_, err := service.Response(testContext, testTaxiRequestMoscow, priceOff)
	assert.NotNil(t, err)
}

func TestAPIGetterErrButHaveResponseFromAPI(t *testing.T) {
	//  Даже есть от APIGetter приходили ошибки, но если есть APIData, то показываем их, игнорируя ошибки
	provName := "test1"
	prod := product.Product{}
	prod.ProviderName = provName
	pcache := newMockProductCache(nil, prod)

	apiDataErr := errors.New("API Data err")
	dg := []APIDataGetter{
		newMockAPIDataGetter(apiDataErr, provName, APIData{}),
	}
	dt := newMockDistanceTimeService(nil, 1000, 2000)
	addr := newMockAddressService(nil, "")

	service := NewBuilder().
		WithAPIs(dg).
		WithProductCache(pcache).
		WithRequester(testHTTPRequester).
		WithDistanceTimeSrv(dt).
		WithAddressSrv(addr).
		WithStatCollector(testCollector).
		WithLogger(testLogger).
		Build()
	_, err := service.Response(testContext, testTaxiRequestMoscow, priceOff)
	assert.Nil(t, err)
}

func TestAPIGetterErrAndNoResponseFromAPI(t *testing.T) {
	//  Даже есть от APIGetter приходили ошибки и одновременно не пришло ни одного APIData, тогда возвращаем ошибку клиенту
	provName := "test1"
	prod := product.Product{}
	prod.ProviderName = provName
	pcache := newMockProductCache(nil, prod)

	apiDataErr := errors.New("API Data err")
	dg := []APIDataGetter{
		newMockAPIDataGetter(apiDataErr, provName),
	}
	dt := newMockDistanceTimeService(nil, 1000, 2000)
	addr := newMockAddressService(nil, "")

	service := NewBuilder().
		WithAPIs(dg).
		WithProductCache(pcache).
		WithRequester(testHTTPRequester).
		WithDistanceTimeSrv(dt).
		WithAddressSrv(addr).
		WithStatCollector(testCollector).
		WithLogger(testLogger).
		Build()
	_, err := service.Response(testContext, testTaxiRequestMoscow, priceOff)
	assert.NotNil(t, err)
}

func TestAPIGetterOKAndNoResponseFromAPI(t *testing.T) {
	//  Даже есть от APIGetter НЕ приходили ошибки, но одновременно не пришло ни одного APIData, тогда возвращаем ошибку клиенту
	provName := "test1"
	prod := product.Product{}
	prod.ProviderName = provName
	pcache := newMockProductCache(nil, prod)

	dg := []APIDataGetter{
		newMockAPIDataGetter(nil, provName),
	}
	dt := newMockDistanceTimeService(nil, 1000, 2000)
	addr := newMockAddressService(nil, "")

	service := NewBuilder().
		WithAPIs(dg).
		WithProductCache(pcache).
		WithRequester(testHTTPRequester).
		WithDistanceTimeSrv(dt).
		WithAddressSrv(addr).
		WithStatCollector(testCollector).
		WithLogger(testLogger).
		Build()
	_, err := service.Response(testContext, testTaxiRequestMoscow, priceOff)
	assert.NotNil(t, err)
}

func TestDistanceTimeErr(t *testing.T) {
	// Неответ от моисея не генерирует ошибку
	dtErr := errors.New("Distance time err")
	dt := newMockDistanceTimeService(dtErr, 1000, 2000)
	service := getTestService()
	service.distTimeSrv = dt
	_, err := service.Response(testContext, testTaxiRequestMoscow, priceOff)
	assert.Nil(t, err)
}

func TestParseReq(t *testing.T) {
	payload := strings.NewReader("{\r\n    \"region_id\": 14, \r\n    \"point1\": { \r\n        \"lon\": 30.723449,\r\n        \"lat\": 46.441982\r\n    },\r\n    \"point2\": { \r\n        \"lon\": 30.759315,\r\n        \"lat\": 46.453352\r\n    },\r\n    \"only_api\": true \r\n}")
	testReq, errTestReq := http.NewRequest("POST", "/taksa/api/1.0/route/calculate", payload)
	if errTestReq != nil {
		t.Fatal(errTestReq)
	}
	service := getTestService()
	req, err := service.ParseTaxiRequest(testContext, testReq)
	service.EvaluateTaxiRequest(testContext, &req)
	assert.Nil(t, err)
	assert.Equal(t, 14, req.RegionID)
	assert.Equal(t, 30.1, req.Point1.Lat)
	assert.Equal(t, "30.1", req.Point1.LatStr)
	assert.Equal(t, 40.0, req.Point1.Lon)
	assert.Equal(t, "40", req.Point1.LonStr)

	assert.Equal(t, 30.1, req.Point2.Lat)
	assert.Equal(t, "30.1", req.Point2.LatStr)
	assert.Equal(t, 40.0, req.Point2.Lon)
	assert.Equal(t, "40", req.Point2.LonStr)
}

func TestParseReqEmptyPointErr(t *testing.T) {
	testReqData := Request{
		RegionID: 1,
		Point1:   Point{},
		Point2: Point{
			Lat: 64.1111,
			Lon: 65.0000,
		},
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(testReqData)
	testReq, _ := http.NewRequest(http.MethodPost, "http://test.com", b)
	service := getTestService()
	_, err := service.ParseTaxiRequest(testContext, testReq)
	assert.NotNil(t, err)
}

func TestParseReqAddrErr(t *testing.T) {
	testReqData := Request{
		RegionID: 1,
		Point1: Point{
			Lat: 54.1111,
			Lon: 55.0000,
		},
		Point2: Point{
			Lat: 64.1111,
			Lon: 65.0000,
		},
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(testReqData)
	testReq, _ := http.NewRequest(http.MethodPost, "http://test.com", b)
	errWebAPI := errors.New("No address in Web API")
	addr := newMockAddressService(errWebAPI, "")
	service := getTestService()
	service.addrSrv = addr
	req, _ := service.ParseTaxiRequest(testContext, testReq)
	err := service.EvaluateTaxiRequest(testContext, &req)
	assert.NotNil(t, err)
}
