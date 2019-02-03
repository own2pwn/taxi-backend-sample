package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/httprequester"
	"github.com/nburunova/taxi-backend-sample/src/taxi/service"
	"gopkg.in/h2non/gock.v1"
)

var tUberAPI = uberAPI{
	Name:        "uber",
	Host:        "https://api.uber.com",
	PriceMethod: "/v1/estimates/price",
	TimeMethod:  "/v1/estimates/time",
	Headers: []httprequester.Dict{
		{
			Key:   "Authorization",
			Value: "Token iyZVcltmUb75Qy55EsNTZ2CHVM9U881AcPHtf4ny",
		},
	},
	DgisClientID: "xxx",
}

func TestUberAPI(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.uber.com").
		Get("/v1/estimates/time").
		Reply(200).
		File("_test_jsons/uberTimes.json")

	gock.New("https://api.uber.com").
		Get("/v1/estimates/price").
		Reply(200).
		File("_test_jsons/uberPrices.json")

	gock.InterceptClient(testHttpClient)
	res, err := tUberAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	assert.Nil(t, err)
	assert.Equal(t, res, []service.APIData{
		service.APIData{
			DisplayName: "Uber BLACK",
			PriceMin:    65,
			PriceMax:    75,
			PriceMean:   70,
			ProductID:   "17962f9a-4392-4260-97b0-d838dc7bb0df",
			TariffName:  "black",
			Eta:         6,
			TemplateVars: map[string]string{
				"%product.id%":   "17962f9a-4392-4260-97b0-d838dc7bb0df",
				"%client.id%":    "xxx",
				"%from.lat%":     testTaxiRequestMoscow.Point1.LatStr,
				"%from.lon%":     testTaxiRequestMoscow.Point1.LonStr,
				"%from.address%": testTaxiRequestMoscow.Point1.Address,
				"%to.lat%":       testTaxiRequestMoscow.Point2.LatStr,
				"%to.lon%":       testTaxiRequestMoscow.Point2.LonStr,
				"%to.address%":   testTaxiRequestMoscow.Point2.Address,
			},
		},
		service.APIData{
			DisplayName: "Uber SELECT",
			PriceMin:    52,
			PriceMax:    60,
			PriceMean:   56,
			ProductID:   "7e7eda63-d2bc-474d-9988-e03bef2320d8",
			TariffName:  "select",
			Eta:         1,
			TemplateVars: map[string]string{
				"%product.id%":   "7e7eda63-d2bc-474d-9988-e03bef2320d8",
				"%client.id%":    "xxx",
				"%from.lat%":     testTaxiRequestMoscow.Point1.LatStr,
				"%from.lon%":     testTaxiRequestMoscow.Point1.LonStr,
				"%from.address%": testTaxiRequestMoscow.Point1.Address,
				"%to.lat%":       testTaxiRequestMoscow.Point2.LatStr,
				"%to.lon%":       testTaxiRequestMoscow.Point2.LonStr,
				"%to.address%":   testTaxiRequestMoscow.Point2.Address,
			},
		},
	})
	assert.Equal(t, gock.IsDone(), true)
}

func TestUberAPITime0(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.uber.com").
		Get("/v1/estimates/time").
		Reply(200).
		File("_test_jsons/uberTimes.0.json")

	gock.New("https://api.uber.com").
		Get("/v1/estimates/price").
		Reply(200).
		File("_test_jsons/uberPrices.json")

	gock.InterceptClient(testHttpClient)
	res, err := tUberAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), service.ErrInvalidTime.Error())
	assert.Equal(t, res, []service.APIData{
		service.APIData{
			DisplayName: "Uber BLACK",
			PriceMin:    65,
			PriceMax:    75,
			PriceMean:   70,
			ProductID:   "17962f9a-4392-4260-97b0-d838dc7bb0df",
			TariffName:  "black",
			Eta:         0,
			TemplateVars: map[string]string{
				"%product.id%":   "17962f9a-4392-4260-97b0-d838dc7bb0df",
				"%client.id%":    "xxx",
				"%from.lat%":     testTaxiRequestMoscow.Point1.LatStr,
				"%from.lon%":     testTaxiRequestMoscow.Point1.LonStr,
				"%from.address%": testTaxiRequestMoscow.Point1.Address,
				"%to.lat%":       testTaxiRequestMoscow.Point2.LatStr,
				"%to.lon%":       testTaxiRequestMoscow.Point2.LonStr,
				"%to.address%":   testTaxiRequestMoscow.Point2.Address,
			},
		},
		service.APIData{
			DisplayName: "Uber SELECT",
			PriceMin:    52,
			PriceMax:    60,
			PriceMean:   56,
			ProductID:   "7e7eda63-d2bc-474d-9988-e03bef2320d8",
			TariffName:  "select",
			Eta:         0,
			TemplateVars: map[string]string{
				"%product.id%":   "7e7eda63-d2bc-474d-9988-e03bef2320d8",
				"%client.id%":    "xxx",
				"%from.lat%":     testTaxiRequestMoscow.Point1.LatStr,
				"%from.lon%":     testTaxiRequestMoscow.Point1.LonStr,
				"%from.address%": testTaxiRequestMoscow.Point1.Address,
				"%to.lat%":       testTaxiRequestMoscow.Point2.LatStr,
				"%to.lon%":       testTaxiRequestMoscow.Point2.LonStr,
				"%to.address%":   testTaxiRequestMoscow.Point2.Address,
			},
		},
	})
	assert.Equal(t, gock.IsDone(), true)
}
func TestUberAPITimeEmpty(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.uber.com").
		Get("/v1/estimates/time").
		Reply(200).
		File("_test_jsons/uberTimes.Empty.json")

	gock.New("https://api.uber.com").
		Get("/v1/estimates/price").
		Reply(200).
		File("_test_jsons/uberPrices.json")

	gock.InterceptClient(testHttpClient)
	res, err := tUberAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), service.ErrInvalidTime.Error())
	assert.Equal(t, res, []service.APIData{
		service.APIData{
			DisplayName: "Uber BLACK",
			PriceMin:    65,
			PriceMax:    75,
			PriceMean:   70,
			ProductID:   "17962f9a-4392-4260-97b0-d838dc7bb0df",
			TariffName:  "black",
			Eta:         0,
			TemplateVars: map[string]string{
				"%product.id%":   "17962f9a-4392-4260-97b0-d838dc7bb0df",
				"%client.id%":    "xxx",
				"%from.lat%":     testTaxiRequestMoscow.Point1.LatStr,
				"%from.lon%":     testTaxiRequestMoscow.Point1.LonStr,
				"%from.address%": testTaxiRequestMoscow.Point1.Address,
				"%to.lat%":       testTaxiRequestMoscow.Point2.LatStr,
				"%to.lon%":       testTaxiRequestMoscow.Point2.LonStr,
				"%to.address%":   testTaxiRequestMoscow.Point2.Address,
			},
		},
		service.APIData{
			DisplayName: "Uber SELECT",
			PriceMin:    52,
			PriceMax:    60,
			PriceMean:   56,
			ProductID:   "7e7eda63-d2bc-474d-9988-e03bef2320d8",
			TariffName:  "select",
			Eta:         0,
			TemplateVars: map[string]string{
				"%product.id%":   "7e7eda63-d2bc-474d-9988-e03bef2320d8",
				"%client.id%":    "xxx",
				"%from.lat%":     testTaxiRequestMoscow.Point1.LatStr,
				"%from.lon%":     testTaxiRequestMoscow.Point1.LonStr,
				"%from.address%": testTaxiRequestMoscow.Point1.Address,
				"%to.lat%":       testTaxiRequestMoscow.Point2.LatStr,
				"%to.lon%":       testTaxiRequestMoscow.Point2.LonStr,
				"%to.address%":   testTaxiRequestMoscow.Point2.Address,
			},
		},
	})
	assert.Equal(t, gock.IsDone(), true)
}

func TestUberAPITimeErr500(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.uber.com").
		Get("/v1/estimates/time").
		Reply(500)

	gock.New("https://api.uber.com").
		Get("/v1/estimates/price").
		Reply(200).
		File("_test_jsons/uberPrices.json")

	gock.InterceptClient(testHttpClient)
	res, err := tUberAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	assert.NotNil(t, err)
	assert.Equal(t, res, []service.APIData{
		service.APIData{
			DisplayName: "Uber BLACK",
			PriceMin:    65,
			PriceMax:    75,
			PriceMean:   70,
			ProductID:   "17962f9a-4392-4260-97b0-d838dc7bb0df",
			TariffName:  "black",
			Eta:         0,
			TemplateVars: map[string]string{
				"%product.id%":   "17962f9a-4392-4260-97b0-d838dc7bb0df",
				"%client.id%":    "xxx",
				"%from.lat%":     testTaxiRequestMoscow.Point1.LatStr,
				"%from.lon%":     testTaxiRequestMoscow.Point1.LonStr,
				"%from.address%": testTaxiRequestMoscow.Point1.Address,
				"%to.lat%":       testTaxiRequestMoscow.Point2.LatStr,
				"%to.lon%":       testTaxiRequestMoscow.Point2.LonStr,
				"%to.address%":   testTaxiRequestMoscow.Point2.Address,
			},
		},
		service.APIData{
			DisplayName: "Uber SELECT",
			PriceMin:    52,
			PriceMax:    60,
			PriceMean:   56,
			ProductID:   "7e7eda63-d2bc-474d-9988-e03bef2320d8",
			TariffName:  "select",
			Eta:         0,
			TemplateVars: map[string]string{
				"%product.id%":   "7e7eda63-d2bc-474d-9988-e03bef2320d8",
				"%client.id%":    "xxx",
				"%from.lat%":     testTaxiRequestMoscow.Point1.LatStr,
				"%from.lon%":     testTaxiRequestMoscow.Point1.LonStr,
				"%from.address%": testTaxiRequestMoscow.Point1.Address,
				"%to.lat%":       testTaxiRequestMoscow.Point2.LatStr,
				"%to.lon%":       testTaxiRequestMoscow.Point2.LonStr,
				"%to.address%":   testTaxiRequestMoscow.Point2.Address,
			},
		},
	})
	assert.Equal(t, gock.IsDone(), true)
}

func TestUberAPINoPrice(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.uber.com").
		Get("/v1/estimates/time").
		Reply(200).
		File("_test_jsons/uberTime.json")

	gock.New("https://api.uber.com").
		Get("/v1/estimates/price").
		Reply(200).
		BodyString("err")

	gock.InterceptClient(testHttpClient)
	res, err := tUberAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(res))
	assert.Equal(t, gock.IsDone(), true)
}

func TestUberAPINoPrice500(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.uber.com").
		Get("/v1/estimates/time").
		Reply(200).
		File("_test_jsons/uberTime.json")

	gock.New("https://api.uber.com").
		Get("/v1/estimates/price").
		Reply(500)

	gock.InterceptClient(testHttpClient)
	res, err := tUberAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(res))
	assert.Equal(t, gock.IsDone(), true)
}

func TestUberAPIPrice0(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.uber.com").
		Get("/v1/estimates/time").
		Reply(200).
		File("_test_jsons/uberTime.json")

	gock.New("https://api.uber.com").
		Get("/v1/estimates/price").
		Reply(200).
		File("_test_jsons/uberPrices.0.json")

	gock.InterceptClient(testHttpClient)
	res, err := tUberAPI.GetAPIData(testContext, testHTTPRequester, testTaxiRequestMoscow)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), service.ErrInvalidPrice.Error())
	assert.Equal(t, 0, len(res))
	assert.Equal(t, gock.IsDone(), true)
}
