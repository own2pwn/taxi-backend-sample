package provider

import (
	"github.com/nburunova/taxi-backend-sample/src/taxi/service"
)

// ProvidersList - список провайдеров
type ProvidersList struct {
	Gett      gettAPI      `json:"gett"`
	Uber      uberAPI      `json:"uber"`
	Citymobil citymobilAPI `json:"citymobil"`
}

// APIGetters - возвращает все провайдеры
func (pl *ProvidersList) APIGetters() []service.APIDataGetter {
	return []service.APIDataGetter{
		pl.Gett,
		pl.Uber,
		pl.Citymobil,
	}
}

func secondsToMins(eta int) int {
	if eta == 0 {
		return 0
	}
	if eta < 60 {
		return 1
	}
	return eta / 60
}
