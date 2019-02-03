package provider

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/httprequester"

	"github.com/nburunova/taxi-backend-sample/src/taxi/service"
)

type uberAPI struct {
	Name         string
	Headers      []httprequester.Dict
	Host         string
	PriceMethod  string
	TimeMethod   string
	DgisClientID string
}

type uberPrice struct {
	LocalizedDisplayName string  `json:"localized_display_name"`
	Distance             float64 `json:"distance"`
	DisplayName          string  `json:"display_name"`
	ProductID            string  `json:"product_id"`
	HighEstimate         float64 `json:"high_estimate"`
	LowEstimate          float64 `json:"low_estimate"`
	Duration             int     `json:"duration"`
	Estimate             string  `json:"estimate"`
	Currency             string  `json:"currency_code"`
	SurgeMultiplier      float64 `json:"surge_multiplier"`
	Minimum              int     `json:"minimum"`
}

type uberPricesResponse struct {
	Prices []uberPrice `json:"prices"`
}

type uberTime struct {
	ProductID   string `json:"product_id"`
	DisplayName string `json:"display_name"`
	Estimate    int    `json:"estimate"`
}

type uberTimesResponse struct {
	Times []uberTime `json:"times"`
}

func (h uberAPI) APIName() string {
	return h.Name
}

func (h uberAPI) GetAPIData(ctx context.Context, httpreq *httprequester.Requester, taxiReq service.Request) ([]service.APIData, error) {
	var errPrices, errTimes error
	var uberPrices *uberPricesResponse
	var uberTimes *uberTimesResponse
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		uberPrices, errPrices = h.prices(ctx, httpreq, taxiReq.Point1, taxiReq.Point2)
	}()
	go func() {
		defer wg.Done()
		uberTimes, errTimes = h.times(ctx, httpreq, taxiReq.Point1)
	}()
	wg.Wait()

	if errPrices != nil {
		return nil, errors.Wrap(errPrices, "Uber: Cannot request prices")
	}
	apiDatas := h.makeAPIDatas(*uberPrices, *uberTimes, taxiReq.Point1, taxiReq.Point2)
	if errTimes != nil {
		return apiDatas, errors.Wrap(errTimes, "Uber: Cannot request times")
	}
	return apiDatas, nil
}

var uberTariffMap = map[string]string{
	"black":      "black",
	"uberblack":  "uberblack",
	"select":     "select",
	"uberselect": "uberselect",

	"one": "one",
	"xl":  "uberxl",
	"x":   "uberx",

	"uberone":     "one",
	"uberxl":      "uberxl",
	"uberx":       "uberx",
	"uberpop":     "uberpop",
	"uberstart":   "uberstart",
	"ubereconomy": "ubereconomy",
}

var uberDisplayMap = map[string]string{
	"black":      "Uber BLACK",
	"select":     "Uber SELECT",
	"one":        "Uber ONE",
	"xl":         "Uber XL",
	"x":          "Uber X",
	"uberblack":  "Uber BLACK",
	"uberselect": "Uber SELECT",
	"uberone":    "Uber ONE",
	"uberxl":     "Uber XL",
	"uberx":      "Uber X",
}

func (h uberAPI) makeAPIDatas(prices uberPricesResponse, uberTimes uberTimesResponse, p1, p2 service.Point) []service.APIData {
	result := make([]service.APIData, 0)
	for _, p := range prices.Prices {
		var pEta int
		for _, e := range uberTimes.Times {
			if e.ProductID == p.ProductID {
				pEta = e.Estimate
				break
			}
		}
		formattedDisplayName := strings.ToLower(p.DisplayName)
		displayName := "Uber"
		if val, ok := uberDisplayMap[formattedDisplayName]; ok {
			displayName = val
		}
		var tariffName string
		if val, ok := uberTariffMap[formattedDisplayName]; ok {
			tariffName = val
		}
		data := service.APIData{
			DisplayName: displayName,
			PriceMax:    int(p.HighEstimate),
			PriceMin:    int(p.LowEstimate),
			PriceMean:   float64(p.HighEstimate+p.LowEstimate) / 2,
			ProductID:   p.ProductID,
			TariffName:  tariffName,
			Eta:         secondsToMins(pEta),
			TemplateVars: map[string]string{
				"%from.lat%":     p1.LatStr,
				"%from.lon%":     p1.LonStr,
				"%from.address%": p1.Address,
				"%to.lat%":       p2.LatStr,
				"%to.lon%":       p2.LonStr,
				"%to.address%":   p2.Address,
				"%product.id%":   p.ProductID,
				"%client.id%":    h.DgisClientID,
			},
		}
		result = append(result, data)
	}
	return result
}

func (h uberAPI) prices(ctx context.Context, httpreq *httprequester.Requester, p1, p2 service.Point) (*uberPricesResponse, error) {
	params := []httprequester.Dict{
		{
			Key:   "start_latitude",
			Value: p1.LatStr,
		},
		{
			Key:   "start_longitude",
			Value: p1.LonStr,
		},
		{
			Key:   "end_latitude",
			Value: p2.LatStr,
		},
		{
			Key:   "end_longitude",
			Value: p2.LonStr,
		},
	}
	prices := new(uberPricesResponse)
	priceURL := fmt.Sprintf("%v%v", h.Host, h.PriceMethod)
	err := httpreq.Get(ctx, priceURL, h.Headers, params, prices)
	if err != nil {
		return prices, err
	}
	if len(prices.Prices) == 0 {
		return prices, errors.Wrap(service.ErrInvalidPrice, "Price list is empty")
	}
	for _, price := range prices.Prices {
		if (price.LowEstimate+price.HighEstimate)/2 <= 0 {
			return prices, errors.Wrap(service.ErrInvalidPrice, "Price <= 0")
		}
	}
	return prices, nil
}

func (h uberAPI) times(ctx context.Context, httpreq *httprequester.Requester, p1 service.Point) (*uberTimesResponse, error) {
	params := []httprequester.Dict{
		{
			Key:   "start_latitude",
			Value: p1.LatStr,
		},
		{
			Key:   "start_longitude",
			Value: p1.LonStr,
		},
	}
	uberTimes := new(uberTimesResponse)
	timeURL := fmt.Sprintf("%v%v", h.Host, h.TimeMethod)
	err := httpreq.Get(ctx, timeURL, h.Headers, params, uberTimes)
	if err != nil {
		return uberTimes, err
	}
	if len(uberTimes.Times) == 0 {
		return uberTimes, errors.Wrap(service.ErrInvalidTime, "Time list is empty")
	}
	for _, e := range uberTimes.Times {
		if e.Estimate <= 0 {
			return uberTimes, errors.Wrap(service.ErrInvalidTime, "Time <= 0")
		}
	}
	return uberTimes, nil
}
