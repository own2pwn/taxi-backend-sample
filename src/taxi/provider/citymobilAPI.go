package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/httprequester"
	"github.com/nburunova/taxi-backend-sample/src/taxi/service"
)

type tariffGroup struct {
	ID   int
	Name string
}

type citymobilAPI struct {
	Name            string
	Host            string
	PriceMethod     string
	PriceMethodName string
	TariffGroups    []tariffGroup
	Ver             string
	Hurry           string
}

type citymobilOrder struct {
	Latitude     string `json:"latitude"`
	Longitude    string `json:"longitude"`
	DelLatitude  string `json:"del_latitude"`
	DelLongitude string `json:"del_longitude"`
	TariffGroup  []int  `json:"tariff_group"`
	Method       string `json:"method"`
	Ver          string `json:"ver"`
	Hurry        string `json:"hurry"`
}

type citymobilPrice struct {
	TariffGroupID int     `json:"id_tariff_group"`
	TotalPrice    float64 `json:"total_price"`
}

type citymobilPriceResponse struct {
	Prices []citymobilPrice `json:"prices"`
}

func (h citymobilAPI) APIName() string {
	return h.Name
}

func (h citymobilAPI) GetAPIData(ctx context.Context, httpreq *httprequester.Requester, taxiReq service.Request) ([]service.APIData, error) {
	tariffGrps := make([]int, len(h.TariffGroups))
	for i, gr := range h.TariffGroups {
		tariffGrps[i] = gr.ID
	}
	priceResp, errPrices := h.price(ctx, httpreq, taxiReq.Point1, taxiReq.Point2, tariffGrps)
	if errPrices != nil {
		return nil, errors.Wrap(errPrices, "citymobil: Cannot request prices")
	}
	apiDatas := h.makeAPIDatas(*priceResp, taxiReq.Point1, taxiReq.Point2)
	return apiDatas, nil
}

func (h citymobilAPI) makeAPIDatas(priceResp citymobilPriceResponse, p1 service.Point, p2 service.Point) []service.APIData {
	result := make([]service.APIData, len(priceResp.Prices))
	for i, priceItem := range priceResp.Prices {
		displayName := "Ситимобил"
		for _, gr := range h.TariffGroups {
			if priceItem.TariffGroupID == gr.ID {
				displayName = "Ситимобил " + gr.Name
			}
		}
		result[i] = service.APIData{
			DisplayName: displayName,
			PriceMean:   priceItem.TotalPrice,
			TemplateVars: map[string]string{
				"%from.lat%":     p1.LatStr,
				"%from.lon%":     p1.LonStr,
				"%from.address%": p1.Address,
				"%to.lat%":       p2.LatStr,
				"%to.lon%":       p2.LonStr,
				"%to.address%":   p2.Address,
			},
		}
	}
	return result
}

func (h citymobilAPI) price(ctx context.Context, httpreq *httprequester.Requester, p1, p2 service.Point, tariffGrps []int) (*citymobilPriceResponse, error) {
	var p = new(citymobilPriceResponse)
	orderData := citymobilOrder{
		Latitude:     p1.LatStr,
		Longitude:    p1.LonStr,
		DelLatitude:  p2.LatStr,
		DelLongitude: p2.LonStr,
		TariffGroup:  tariffGrps,
		Method:       h.PriceMethodName,
		Ver:          h.Ver,
		Hurry:        h.Hurry,
	}
	bData, errMarsh := json.Marshal(orderData)
	if errMarsh != nil {
		return p, errMarsh
	}
	priceURL := fmt.Sprintf("%v%v", h.Host, h.PriceMethod)
	errPrice := httpreq.Post(ctx, priceURL, []httprequester.Dict{}, bData, p)
	if errPrice != nil {
		return p, errPrice
	}
	if len(p.Prices) == 0 {
		return p, errors.Wrap(service.ErrInvalidPrice, "Price list empty")
	}
	for _, priceItem := range p.Prices {
		if priceItem.TotalPrice <= 0 {
			return p, errors.Wrap(service.ErrInvalidPrice, "Price <= 0")
		}
	}
	return p, nil
}
