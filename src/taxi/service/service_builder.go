package service

import (
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/collector"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/httprequester"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/log"
)

// Builder - создатель сервиса Таксы
type Builder struct {
	apis        []APIDataGetter
	prodCache   ProductsCache
	httpreq     *httprequester.Requester
	distTimeSrv DistanceTimeService
	addrSrv     AddressService
	collector   *collector.Collector
	logger      *log.StructuredLogger
}

// NewBuilder - создаем создателя сервиса Таксы
func NewBuilder() *Builder {
	return new(Builder)
}

// WithAPIs - передаем список АПИ такси в сервис Таксы
func (sb *Builder) WithAPIs(dgetters []APIDataGetter) *Builder {
	sb.apis = dgetters
	return sb
}

// WithProductCache - передаем кэш с продуктами в сервис Таксы
func (sb *Builder) WithProductCache(pc ProductsCache) *Builder {
	sb.prodCache = pc
	return sb
}

// WithRequester - передаем HTTP клиент в сервис Таксы
func (sb *Builder) WithRequester(hr *httprequester.Requester) *Builder {
	sb.httpreq = hr
	return sb
}

// WithDistanceTimeSrv - передаем сервис, который возврашает время проезда и расстояние между точками А и Б
func (sb *Builder) WithDistanceTimeSrv(dts DistanceTimeService) *Builder {
	sb.distTimeSrv = dts
	return sb
}

// WithAddressSrv - передаем сервис, который возвращает адрес и принадлжность области для гео-точки
func (sb *Builder) WithAddressSrv(as AddressService) *Builder {
	sb.addrSrv = as
	return sb
}

// WithStatCollector - передаем сборшик статистики в сервис Таксы
func (sb *Builder) WithStatCollector(c *collector.Collector) *Builder {
	sb.collector = c
	return sb
}

// WithLogger - передаем логгер в сервис таксы
func (sb *Builder) WithLogger(l *log.StructuredLogger) *Builder {
	sb.logger = l
	return sb
}

// Build - создаем сервис таксы со всеми переданными данными
func (sb *Builder) Build() *Service {
	apisMap := make(map[string]APIDataGetter)
	for _, api := range sb.apis {
		apisMap[api.APIName()] = api
	}
	s := Service{
		apisMap:     apisMap,
		prodCache:   sb.prodCache,
		httpreq:     sb.httpreq,
		distTimeSrv: sb.distTimeSrv,
		addrSrv:     sb.addrSrv,
		collector:   sb.collector,
		Logger:      sb.logger,
	}
	if s.IsOK() {
		return &s
	}
	return nil
}
