package provider

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/nburunova/taxi-backend-sample/src/taxi/service"
	"gopkg.in/h2non/gock.v1"
)

var tCitymobilAPI = citymobilAPI{
	Name:            "citymobil",
	Host:            "https://c-api.city-mobil.ru",
	PriceMethod:     "/",
	PriceMethodName: "getprice",
	TariffGroups: []tariffGroup{
		tariffGroup{
			ID:   2,
			Name: "Эконом",
		},
		tariffGroup{
			ID:   4,
			Name: "Не Эконом",
		},
	},
	Ver:   "4.0.0",
	Hurry: "1",
}

func TestCitymobilAPI(t *testing.T) {
	defer gock.Off()
	gock.New(tCitymobilAPI.Host).
		Post(tCitymobilAPI.PriceMethod).
		Reply(200).
		File("_test_jsons/citymobil.json")

	gock.InterceptClient(testHttpClient)
	res, err := tCitymobilAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	sort.Sort(service.ByProductID(res))
	assert.Nil(t, err)
	assert.Equal(t, []service.APIData{
		service.APIData{
			DisplayName: "Ситимобил Эконом",
			PriceMean:   449.0,
			TemplateVars: map[string]string{
				"%from.lat%":     testTaxiRequestMoscow.Point1.LatStr,
				"%from.lon%":     testTaxiRequestMoscow.Point1.LonStr,
				"%from.address%": testTaxiRequestMoscow.Point1.Address,
				"%to.lat%":       testTaxiRequestMoscow.Point2.LatStr,
				"%to.lon%":       testTaxiRequestMoscow.Point2.LonStr,
				"%to.address%":   testTaxiRequestMoscow.Point2.Address,
			},
		},
	}, res)
	assert.Equal(t, gock.IsDone(), true)
}

func TestCitymobilAPIPrice500(t *testing.T) {
	defer gock.Off()
	gock.New(tCitymobilAPI.Host).
		Post(tCitymobilAPI.PriceMethod).
		Reply(500)

	gock.InterceptClient(testHttpClient)
	res, err := tCitymobilAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	sort.Sort(service.ByProductID(res))
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(res))
	assert.Equal(t, gock.IsDone(), true)
}

func TestCitymobilAPIPrice0(t *testing.T) {
	defer gock.Off()
	gock.New(tCitymobilAPI.Host).
		Post(tCitymobilAPI.PriceMethod).
		Reply(200).
		File("_test_jsons/citymobil.0.json")

	gock.InterceptClient(testHttpClient)
	res, err := tCitymobilAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	sort.Sort(service.ByProductID(res))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), service.ErrInvalidPrice.Error())
	assert.Equal(t, 0, len(res))
	assert.Equal(t, gock.IsDone(), true)
}

func TestCitymobilAPINone(t *testing.T) {
	defer gock.Off()
	gock.New(tCitymobilAPI.Host).
		Post(tCitymobilAPI.PriceMethod).
		Reply(200).
		File("_test_jsons/citymobil.none.json")

	gock.InterceptClient(testHttpClient)
	res, err := tCitymobilAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	sort.Sort(service.ByProductID(res))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), service.ErrInvalidPrice.Error())
	assert.Equal(t, 0, len(res))
	assert.Equal(t, gock.IsDone(), true)
}

func TestCitymobilAPINoPrice(t *testing.T) {
	defer gock.OffAll()
	gock.New(tCitymobilAPI.Host).
		Post(tCitymobilAPI.PriceMethod).
		Reply(200).
		BodyString("err")

	gock.InterceptClient(testHttpClient)
	res, err := tCitymobilAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(res))
}
