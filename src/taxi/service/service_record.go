package service

import (
	"github.com/nburunova/taxi-backend-sample/src/product"
)

type serviceRecord struct {
	AvgEta       *int             `json:"avg_eta"`
	Price        int              `json:"price"`
	PriceRanges  *priceRanges     `json:"price_ranges"`
	Rating       *float32         `json:"rating"`
	Operator     product.Operator `json:"operator"`
	Eta          *int             `json:"eta"`
	CurrencyCode *string          `json:"currency_code"`
}

func newServiceRecord(apiData APIData, prod product.Product) (serviceRecord, error) {
	var eta *int
	if apiData.Eta > 0 {
		etaMins := apiData.Eta
		eta = &etaMins
	} else {
		eta = nil
	}

	var rating *float32
	if prod.Rating != nil && *prod.Rating > 0 {
		rating = prod.Rating
	}

	operator, err := prod.GetOperator(apiData.DisplayName, apiData.TemplateVars)
	return serviceRecord{
		AvgEta:       prod.AvgEta,
		Eta:          eta,
		Price:        int(apiData.PriceMean),
		PriceRanges:  newPriceRanges(apiData.PriceMin, apiData.PriceMax),
		Rating:       rating,
		CurrencyCode: prod.CurrencyCode,
		Operator:     operator,
	}, err
}

type byPrice []serviceRecord

func (a byPrice) Len() int           { return len(a) }
func (a byPrice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPrice) Less(i, j int) bool { return a[i].Price < a[j].Price }

type priceRanges struct {
	Min *int `json:"min"`
	Max *int `json:"max"`
}

func newPriceRanges(min int, max int) *priceRanges {
	if min > 0 && max > 0 {
		prange := priceRanges{
			Min: &min,
			Max: &max,
		}
		return &prange
	}
	return nil
}
