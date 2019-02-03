package collector

import (
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/log"
)

var (
	// ErrCollectorNotFound - не смогли найти собиратель для статистики
	ErrCollectorNotFound = errors.New("Stat Collector not found")
)

// Collector - собиратель информации для прометея
type Collector struct {
	// Количество запросов к провайдерам
	providerRequest *prometheus.CounterVec
	// Количество ответов провайдеров с ошибками 500, 400 и тд
	providerError *prometheus.CounterVec
	// Количество ответов провайдеров, содержащие невалидные данные
	providerInvalidValue *prometheus.CounterVec
	// Количество ответов от провайдеров, в которых все ок
	providerOK *prometheus.CounterVec
	//  Время получения ответа от провайдера
	providerRespTime *prometheus.HistogramVec
	// Количество запросов, отвалившихся по таймауту
	timeout *prometheus.CounterVec
	// Количество ошибочных ответов от сервиса - нет данных от провайдера или от Моисея \ вебапи
	serviceError *prometheus.CounterVec
	// Количество случаев, когда отфильтровали все результаты из ответа провайдера (не подошел тариф)
	filterError *prometheus.CounterVec
	// Последнее успешное обновление кэша
	cacheReload prometheus.Gauge
	// Последнее успешное обновление кодов регионов
	regionsReload prometheus.Gauge
	logger        log.Logger
}

// NewCollector - создать собиратель информации о прометее
func NewCollector(logger log.Logger) *Collector {
	col := Collector{
		providerRequest: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: "navi_taxa_providers",
				Name:      "request",
				Help:      "Количество запросов к провайдерам",
			}, []string{"name", "region"}),
		providerError: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: "navi_taxa_providers",
				Name:      "error_response",
				Help:      "Количество ответов провайдеров с ошибками 500, 400 и тд",
			}, []string{"name", "method", "code", "region"}),
		providerInvalidValue: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: "navi_taxa_providers",
				Name:      "invalid_response",
				Help:      "Количество ответов провайдеров с невалидными данными",
			}, []string{"name", "cause", "region"}),
		providerOK: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: "navi_taxa_providers",
				Name:      "ok_response",
				Help:      "Количество хороших ответов от провайдеров",
			}, []string{"name", "region"}),
		providerRespTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: "navi_taxa_providers",
				Name:      "response_time",
				Help:      "Распределение времени ответов от провайдера",
				Buckets:   prometheus.LinearBuckets(200, 200, 6), // 6 buckets, each 200 ms wide, start 200.
			}, []string{"name", "method"}),
		timeout: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: "navi_taxa_providers",
				Name:      "request_timeout",
				Help:      "Количество запросов, прервавшихся по таймауту",
			}, []string{"name", "region"}),
		serviceError: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: "navi_taxa_service",
				Name:      "error_response",
				Help:      "Количество плохих ответов от сервиса",
			}, []string{"cause", "region"}),
		filterError: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: "navi_taxa_service",
				Name:      "filter_empty",
				Help:      "Количество случаев, когда отфильтровали все результаты из ответа провайдера (не подошел тариф)",
			}, []string{"name", "region"}),
		cacheReload: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Subsystem: "navi_taxa_service",
				Name:      "cache_reload",
				Help:      "Дата последнего обновления кэша",
			},
		),
		regionsReload: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Subsystem: "navi_taxa_service",
				Name:      "regions_reload",
				Help:      "Дата последнего обновления списка кодов регионов",
			},
		),
		logger: logger,
	}
	return &col
}

// RegisterCollections - зарегистрировать все коллекторы статистики в Прометее
func (c *Collector) RegisterCollections() error {
	if err := prometheus.Register(c.providerRequest); err != nil {
		return errors.Wrap(err, "providerRequest")
	}
	if err := prometheus.Register(c.providerError); err != nil {
		return errors.Wrap(err, "providerError")
	}
	if err := prometheus.Register(c.providerInvalidValue); err != nil {
		return errors.Wrap(err, "providerInvalidValue")
	}
	if err := prometheus.Register(c.providerOK); err != nil {
		return errors.Wrap(err, "providerOK")
	}
	if err := prometheus.Register(c.timeout); err != nil {
		return errors.Wrap(err, "timeout")
	}
	if err := prometheus.Register(c.providerRespTime); err != nil {
		return errors.Wrap(err, "providerRespTime")
	}
	if err := prometheus.Register(c.serviceError); err != nil {
		return errors.Wrap(err, "serviceError")
	}
	if err := prometheus.Register(c.filterError); err != nil {
		return errors.Wrap(err, "filterError")
	}
	if err := prometheus.Register(c.cacheReload); err != nil {
		return errors.Wrap(err, "cacheReload")
	}
	if err := prometheus.Register(c.regionsReload); err != nil {
		return errors.Wrap(err, "regionsReload")
	}
	return nil
}

// AddProviderRequest - зарегистрировать запрос к провайдеру
func (c *Collector) AddProviderRequest(providerName string, region int) error {
	counter, err := c.providerRequest.GetMetricWithLabelValues(providerName, strconv.Itoa(region))
	if err != nil {
		c.logger.Error("Provider request collector not found:", err)
		return err
	}
	if counter == nil {
		c.logger.Error("Provider request collector is nil.")
		return ErrCollectorNotFound
	}
	counter.Inc()
	return nil
}

// AddProviderErrorResponse - зарегистрировать ошибку от провайдера
func (c *Collector) AddProviderErrorResponse(providerName, method string, errCode, region int) error {
	counter, err := c.providerError.GetMetricWithLabelValues(providerName, method, strconv.Itoa(errCode), strconv.Itoa(region))
	if err != nil {
		c.logger.Error("Provider errors collector not found:", err)
		return err
	}
	if counter == nil {
		c.logger.Error("Provider errors collector is nil.")
		return ErrCollectorNotFound
	}
	counter.Inc()
	return nil
}

// AddProviderInvalidValueResponse - зарегистрировать не подходящий по формату ответ от провайдера
func (c *Collector) AddProviderInvalidValueResponse(providerName, dataType string, region int) error {
	counter, err := c.providerInvalidValue.GetMetricWithLabelValues(providerName, dataType, strconv.Itoa(region))
	if err != nil {
		c.logger.Error("Provider invalid response collector not found:", err)
		return err
	}
	if counter == nil {
		c.logger.Error("Provider invalid response collector is nil.")
		return ErrCollectorNotFound
	}
	counter.Inc()
	return nil
}

// AddFilterError - зарегистрировать случай, когда отфильтровали все результаты из ответа провайдера
func (c *Collector) AddFilterError(providerName string, region int) error {
	counter, err := c.filterError.GetMetricWithLabelValues(providerName, strconv.Itoa(region))
	if err != nil {
		c.logger.Error("Filter error collector not found:", err)
		return err
	}
	if counter == nil {
		c.logger.Error("Filter error collector is nil.")
		return ErrCollectorNotFound
	}
	counter.Inc()
	return nil
}

// AddProviderOKResponse - зарегистрировать хороший ответ от провайдера
func (c *Collector) AddProviderOKResponse(providerName string, region int) error {
	counter, err := c.providerOK.GetMetricWithLabelValues(providerName, strconv.Itoa(region))
	if err != nil {
		c.logger.Error("Provider OK response collector not found:", err)
		return err
	}
	if counter == nil {
		c.logger.Error("Provider OK response collector is nil.")
		return ErrCollectorNotFound
	}
	counter.Inc()
	return nil
}

// AddRequestTimeout - зарегистрировать запрос к провайдеру, который завершился по таймауту
func (c *Collector) AddRequestTimeout(providerName string, region int) error {
	counter, err := c.timeout.GetMetricWithLabelValues(providerName, strconv.Itoa(region))
	if err != nil {
		c.logger.Error("Provider timeout request collector not found:", err)
		return err
	}
	if counter == nil {
		c.logger.Error("Provider timeout request collector is nil.")
		return ErrCollectorNotFound
	}
	counter.Inc()
	return nil
}

// AddServiceError - зарегистрировать ошибку сервиса
func (c *Collector) AddServiceError(causeName string, region int) error {
	counter, err := c.serviceError.GetMetricWithLabelValues(causeName, strconv.Itoa(region))
	if err != nil {
		c.logger.Error("Service errors collector not found:", err)
		return err
	}
	if counter == nil {
		c.logger.Error("Service errors collector is nil.")
		return ErrCollectorNotFound
	}
	counter.Inc()
	return nil
}

// AddProviderResponseTime - зарегистрировать время ответа от провайдера
func (c *Collector) AddProviderResponseTime(providerName, method string, ms float64) error {
	counter, err := c.providerRespTime.GetMetricWithLabelValues(providerName, method)
	if err != nil {
		c.logger.Error("Service errors collector not found:", err)
		return err
	}
	if counter == nil {
		c.logger.Error("Service errors collector is nil.")
		return ErrCollectorNotFound
	}
	counter.Observe(ms)
	return nil
}

// UpdateCacheReload - обновить время обновления кэша на текущее
func (c *Collector) UpdateCacheReload() error {
	if c.cacheReload == nil {
		c.logger.Error("Cache reload collector is nil.")
		return ErrCollectorNotFound
	}
	c.cacheReload.Set(float64(time.Now().UnixNano()) / 1e9)
	return nil
}

// UpdateRegionsReload - обновить время обновления списка регионов на текущее
func (c *Collector) UpdateRegionsReload() error {
	if c.regionsReload == nil {
		c.logger.Error("Cache reload collector is nil.")
		return ErrCollectorNotFound
	}
	c.regionsReload.Set(float64(time.Now().UnixNano()) / 1e9)
	return nil
}
