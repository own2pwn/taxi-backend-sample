package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/httprequester"
	"github.com/nburunova/taxi-backend-sample/src/taxi/service"
	"gopkg.in/h2non/gock.v1"
)

var tGettAPI = gettAPI{
	Name:        "gett",
	Host:        "https://api.gett.com",
	PriceMethod: "/v1/availability/price",
	TimeMethod:  "/v1/availability/eta",
	Headers: []httprequester.Dict{
		{
			Key:   "Authorization",
			Value: "Token ba02a300d8c1e00c475a6152fe0df4e92cc3ace7ae6a23a297a9f99e08124522",
		},
	},
}

func TestGettAPI(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.gett.com").
		Get("/v1/availability/price").
		Reply(200).
		File("_test_jsons/gettPrice.json")

	gock.New("https://api.gett.com").
		Get("/v1/availability/eta").
		Reply(200).
		File("_test_jsons/gettEta.json")

	gock.InterceptClient(testHttpClient)
	res, err := tGettAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	assert.Nil(t, err)
	assert.Equal(t, res, []service.APIData{
		service.APIData{
			DisplayName: "Gett Комфорт",
			PriceMax:    300,
			PriceMin:    100,
			PriceMean:   200,
			ProductID:   "e9e71379-7d5a-4930-899c-996b83616f87",
			TariffName:  "gett_comfort",
			Eta:         25,
			TemplateVars: map[string]string{
				"%from.lat%":   testTaxiRequestMoscow.Point1.LatStr,
				"%from.lon%":   testTaxiRequestMoscow.Point1.LonStr,
				"%to.lat%":     testTaxiRequestMoscow.Point2.LatStr,
				"%to.lon%":     testTaxiRequestMoscow.Point2.LonStr,
				"%product.id%": "e9e71379-7d5a-4930-899c-996b83616f87",
			},
		},
		service.APIData{
			DisplayName: "Gett Стандарт",
			PriceMax:    100,
			PriceMin:    50,
			PriceMean:   75,
			ProductID:   "6e31c74c-388b-4b84-a362-5048cfc6aa98",
			TariffName:  "gett_standart",
			Eta:         0,
			TemplateVars: map[string]string{
				"%from.lat%":   testTaxiRequestMoscow.Point1.LatStr,
				"%from.lon%":   testTaxiRequestMoscow.Point1.LonStr,
				"%to.lat%":     testTaxiRequestMoscow.Point2.LatStr,
				"%to.lon%":     testTaxiRequestMoscow.Point2.LonStr,
				"%product.id%": "6e31c74c-388b-4b84-a362-5048cfc6aa98",
			},
		},
	})
	assert.Equal(t, gock.IsDone(), true)
}

func TestGettAPINoEta500(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.gett.com").
		Get("/v1/availability/price").
		Reply(200).
		File("_test_jsons/gettPrice.json")

	gock.New("https://api.gett.com").
		Get("/v1/availability/eta").
		Reply(500)

	gock.InterceptClient(testHttpClient)
	res, err := tGettAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	assert.NotNil(t, err)
	assert.Equal(t, res, []service.APIData{
		service.APIData{
			DisplayName: "Gett Комфорт",
			PriceMax:    300,
			PriceMin:    100,
			PriceMean:   200,
			ProductID:   "e9e71379-7d5a-4930-899c-996b83616f87",
			TariffName:  "gett_comfort",
			Eta:         0,
			TemplateVars: map[string]string{
				"%from.lat%":   testTaxiRequestMoscow.Point1.LatStr,
				"%from.lon%":   testTaxiRequestMoscow.Point1.LonStr,
				"%to.lat%":     testTaxiRequestMoscow.Point2.LatStr,
				"%to.lon%":     testTaxiRequestMoscow.Point2.LonStr,
				"%product.id%": "e9e71379-7d5a-4930-899c-996b83616f87",
			},
		},
		service.APIData{
			DisplayName: "Gett Стандарт",
			PriceMax:    100,
			PriceMin:    50,
			PriceMean:   75,
			ProductID:   "6e31c74c-388b-4b84-a362-5048cfc6aa98",
			TariffName:  "gett_standart",
			Eta:         0,
			TemplateVars: map[string]string{
				"%from.lat%":   testTaxiRequestMoscow.Point1.LatStr,
				"%from.lon%":   testTaxiRequestMoscow.Point1.LonStr,
				"%to.lat%":     testTaxiRequestMoscow.Point2.LatStr,
				"%to.lon%":     testTaxiRequestMoscow.Point2.LonStr,
				"%product.id%": "6e31c74c-388b-4b84-a362-5048cfc6aa98",
			},
		},
	})
	assert.Equal(t, gock.IsDone(), true)
}

func TestGettAPIEtaErr(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.gett.com").
		Get("/v1/availability/price").
		Reply(200).
		File("_test_jsons/gettPrice.json")

	gock.New("https://api.gett.com").
		Get("/v1/availability/eta").
		Reply(200).
		BodyString("error")

	gock.InterceptClient(testHttpClient)
	res, err := tGettAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	assert.NotNil(t, err)
	assert.Equal(t, res, []service.APIData{
		service.APIData{
			DisplayName: "Gett Комфорт",
			PriceMax:    300,
			PriceMin:    100,
			PriceMean:   200,
			ProductID:   "e9e71379-7d5a-4930-899c-996b83616f87",
			TariffName:  "gett_comfort",
			Eta:         0,
			TemplateVars: map[string]string{
				"%from.lat%":   testTaxiRequestMoscow.Point1.LatStr,
				"%from.lon%":   testTaxiRequestMoscow.Point1.LonStr,
				"%to.lat%":     testTaxiRequestMoscow.Point2.LatStr,
				"%to.lon%":     testTaxiRequestMoscow.Point2.LonStr,
				"%product.id%": "e9e71379-7d5a-4930-899c-996b83616f87",
			},
		},
		service.APIData{
			DisplayName: "Gett Стандарт",
			PriceMax:    100,
			PriceMin:    50,
			PriceMean:   75,
			ProductID:   "6e31c74c-388b-4b84-a362-5048cfc6aa98",
			TariffName:  "gett_standart",
			Eta:         0,
			TemplateVars: map[string]string{
				"%from.lat%":   testTaxiRequestMoscow.Point1.LatStr,
				"%from.lon%":   testTaxiRequestMoscow.Point1.LonStr,
				"%to.lat%":     testTaxiRequestMoscow.Point2.LatStr,
				"%to.lon%":     testTaxiRequestMoscow.Point2.LonStr,
				"%product.id%": "6e31c74c-388b-4b84-a362-5048cfc6aa98",
			},
		},
	})
	assert.Equal(t, gock.IsDone(), true)
}

func TestGettAPIEtaEmpty(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.gett.com").
		Get("/v1/availability/price").
		Reply(200).
		File("_test_jsons/gettPrice.json")

	gock.New("https://api.gett.com").
		Get("/v1/availability/eta").
		Reply(200).
		File("_test_jsons/gettEtaEmpty.json")

	gock.InterceptClient(testHttpClient)
	res, err := tGettAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), service.ErrInvalidTime.Error())
	assert.Equal(t, res, []service.APIData{
		service.APIData{
			DisplayName: "Gett Комфорт",
			PriceMax:    300,
			PriceMin:    100,
			PriceMean:   200,
			ProductID:   "e9e71379-7d5a-4930-899c-996b83616f87",
			TariffName:  "gett_comfort",
			Eta:         0,
			TemplateVars: map[string]string{
				"%from.lat%":   testTaxiRequestMoscow.Point1.LatStr,
				"%from.lon%":   testTaxiRequestMoscow.Point1.LonStr,
				"%to.lat%":     testTaxiRequestMoscow.Point2.LatStr,
				"%to.lon%":     testTaxiRequestMoscow.Point2.LonStr,
				"%product.id%": "e9e71379-7d5a-4930-899c-996b83616f87",
			},
		},
		service.APIData{
			DisplayName: "Gett Стандарт",
			PriceMax:    100,
			PriceMin:    50,
			PriceMean:   75,
			ProductID:   "6e31c74c-388b-4b84-a362-5048cfc6aa98",
			TariffName:  "gett_standart",
			Eta:         0,
			TemplateVars: map[string]string{
				"%from.lat%":   testTaxiRequestMoscow.Point1.LatStr,
				"%from.lon%":   testTaxiRequestMoscow.Point1.LonStr,
				"%to.lat%":     testTaxiRequestMoscow.Point2.LatStr,
				"%to.lon%":     testTaxiRequestMoscow.Point2.LonStr,
				"%product.id%": "6e31c74c-388b-4b84-a362-5048cfc6aa98",
			},
		},
	})
	assert.Equal(t, gock.IsDone(), true)
}

func TestGettAPINoPrice(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.gett.com").
		Get("/v1/availability/price").
		Reply(200).
		BodyString("error")

	gock.New("https://api.gett.com").
		Get("/v1/availability/eta").
		Reply(200).
		File("_test_jsons/gettEta.json")

	gock.InterceptClient(testHttpClient)
	res, err := tGettAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(res))
	assert.Equal(t, gock.IsDone(), true)
}

func TestGettAPINoPrice500(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.gett.com").
		Get("/v1/availability/price").
		Reply(500)

	gock.New("https://api.gett.com").
		Get("/v1/availability/eta").
		Reply(200).
		File("_test_jsons/gettEta.json")

	gock.InterceptClient(testHttpClient)
	res, err := tGettAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(res))
	assert.Equal(t, gock.IsDone(), true)
}

func TestGettAPIPrice0(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.gett.com").
		Get("/v1/availability/price").
		Reply(200).
		File("_test_jsons/gettPrice.0.json")

	gock.New("https://api.gett.com").
		Get("/v1/availability/eta").
		Reply(200).
		File("_test_jsons/gettEta.json")

	gock.InterceptClient(testHttpClient)
	res, err := tGettAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), service.ErrInvalidPrice.Error())
	assert.Equal(t, 0, len(res))
	assert.Equal(t, gock.IsDone(), true)
}
