package provider

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/nburunova/taxi-backend-sample/src/infrastructure/httprequester"

	"github.com/pkg/errors"
	"github.com/nburunova/taxi-backend-sample/src/taxi/service"
)

type gettAPI struct {
	Name        string
	Headers     []httprequester.Dict
	Host        string
	PriceMethod string
	TimeMethod  string
}

type gettPrice struct {
	Prices []struct {
		ProductID    string `json:"product_id"`
		DisplayName  string `json:"display_name"`
		Estimate     string `json:"estimate"`
		Currency     string `json:"currency"`
		LowEstimate  int    `json:"low_estimate"`
		HighEstimate int    `json:"high_estimate"`
	} `json:"prices"`
}

type gettEta struct {
	Etas []struct {
		ProductID   string `json:"product_id"`
		DisplayName string `json:"display_name"`
		Eta         int    `json:"eta"`
	} `json:"etas"`
}

func (h gettAPI) APIName() string {
	return h.Name
}

func (h gettAPI) GetAPIData(ctx context.Context, httpreq *httprequester.Requester, taxiReq service.Request) ([]service.APIData, error) {
	var errPrices, errEtas error
	var price *gettPrice
	var eta *gettEta
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		price, errPrices = h.price(ctx, httpreq, taxiReq.Point1, taxiReq.Point2)
	}()
	go func() {
		defer wg.Done()
		eta, errEtas = h.time(ctx, httpreq, taxiReq.Point1)
	}()
	wg.Wait()

	if errPrices != nil {
		return nil, errors.Wrap(errPrices, "Gett: Cannot request prices")
	}
	apiDatas := h.makeAPIDatas(*price, *eta, taxiReq.Point1, taxiReq.Point2)
	if errEtas != nil {
		return apiDatas, errors.Wrap(errEtas, "Gett: Cannot request eta")
	}
	return apiDatas, nil
}

func (h gettAPI) makeAPIDatas(price gettPrice, eta gettEta, p1 service.Point, p2 service.Point) []service.APIData {
	result := make([]service.APIData, 0)
	for _, p := range price.Prices {
		var pEta int
		for _, e := range eta.Etas {
			if strings.ToLower(e.DisplayName) == strings.ToLower(p.DisplayName) {
				pEta = e.Eta
				break
			}
		}
		data := service.APIData{
			DisplayName: fmt.Sprintf("Gett %v", p.DisplayName),
			PriceMax:    p.HighEstimate,
			PriceMin:    p.LowEstimate,
			PriceMean:   float64(p.HighEstimate+p.LowEstimate) / 2,
			ProductID:   p.ProductID,
			TariffName:  gettTariffMap[p.DisplayName],
			Eta:         secondsToMins(pEta),
			TemplateVars: map[string]string{
				"%from.lat%":   p1.LatStr,
				"%from.lon%":   p1.LonStr,
				"%to.lat%":     p2.LatStr,
				"%to.lon%":     p2.LonStr,
				"%product.id%": p.ProductID,
			},
		}
		result = append(result, data)
	}
	return result
}

var gettTariffMap = map[string]string{
	"Эконом":      "gett_economy",
	"Комфорт":     "gett_comfort",
	"Бизнес":      "gett_business",
	"Минимум":     "gett_mini",
	"Эконом+":     "gett_economy_plus",
	"Подмосковье": "gett_premoscow",
	"Стандарт":    "gett_standart",
}

func (h gettAPI) price(ctx context.Context, httpreq *httprequester.Requester, p1 service.Point, p2 service.Point) (*gettPrice, error) {
	params := []httprequester.Dict{
		{
			Key:   "pickup_latitude",
			Value: p1.LatStr,
		},
		{
			Key:   "pickup_longitude",
			Value: p1.LonStr,
		},
		{
			Key:   "destination_latitude",
			Value: p2.LatStr,
		},
		{
			Key:   "destination_longitude",
			Value: p2.LonStr,
		},
	}
	var p = new(gettPrice)
	priceURL := fmt.Sprintf("%v%v", h.Host, h.PriceMethod)
	errPrice := httpreq.Get(ctx, priceURL, h.Headers, params, p)
	if errPrice != nil {
		return p, errPrice
	}
	if len(p.Prices) == 0 {
		return p, errors.Wrap(service.ErrInvalidPrice, "Price list is empty")
	}
	for _, price := range p.Prices {
		if float64(price.HighEstimate+price.LowEstimate)/2 <= 0 {
			return p, errors.Wrap(service.ErrInvalidPrice, "Price <= 0")
		}
	}
	return p, nil
}

func (h gettAPI) time(ctx context.Context, httpreq *httprequester.Requester, p1 service.Point) (*gettEta, error) {
	params := []httprequester.Dict{
		{
			Key:   "latitude",
			Value: p1.LatStr,
		},
		{
			Key:   "longitude",
			Value: p1.LonStr,
		},
	}
	var eta = new(gettEta)
	timeURL := fmt.Sprintf("%v%v", h.Host, h.TimeMethod)
	errTime := httpreq.Get(ctx, timeURL, h.Headers, params, eta)
	if errTime != nil {
		return eta, errTime
	}
	if len(eta.Etas) == 0 {
		return eta, errors.Wrap(service.ErrInvalidTime, "Time list is empty")
	}
	for _, e := range eta.Etas {
		if e.Eta <= 0 {
			return eta, errors.Wrap(service.ErrInvalidTime, "Time <= 0")
		}
	}
	return eta, nil
}
