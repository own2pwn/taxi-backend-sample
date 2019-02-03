package settings

import (
	"time"

	"github.com/nburunova/taxi-backend-sample/src/taxi"
	"github.com/nburunova/taxi-backend-sample/src/taxi/provider"
)

// Settings - структура, содержащая настройки и данные для АПИ провайдеов такси
type Settings struct {
	ReloadDBSchedule      string                 `json:"reload_cache_period_cron"`
	ReloadRegionsSchedule string                 `json:"reload_regions_period_cron"`
	WaitTime              time.Duration          `json:"wait_time_ms"`
	PriceCoeff            float64                `json:"price_coeff"`
	RegPriceCoeff         taxi.RegionPriceCoeff  `json:"region_price_coeff"`
	ConnStr               string                 `json:"conn_str"`
	Providers             provider.ProvidersList `json:"taxi_services"`
}

// IsEmply - проверяем, есть ли что-нибудь в настройках
func (s *Settings) IsEmply() bool {
	return s.ConnStr == ""
}
