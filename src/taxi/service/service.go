package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/pkg/errors"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/collector"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/httprequester"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/log"
	"github.com/nburunova/taxi-backend-sample/src/product"
	"github.com/nburunova/taxi-backend-sample/src/webapi"
)

var (
	// ErrNoAPIData - нет данных от провайдера
	ErrNoAPIData = errors.New("No data from providers for this taxi request")
	// ErrTaxiReqEmpty - запрос к сервису пустой
	ErrTaxiReqEmpty = errors.New("Taxi request point is empty")
	// ErrInvalidPrice - in case if price <= 0 or None
	ErrInvalidPrice = errors.New("Invalid price")
	// ErrInvalidTime - in case if price <= 0 or None
	ErrInvalidTime = errors.New("Invalid time")
	// ErrRegionOutOfService - провайдер не обслуживает запрошенный регион
	ErrRegionOutOfService = errors.New("Provider does not service this region")
)

// APIData - cтрутура описывает, в каком формате ожидаем ответ от провайдеров
type APIData struct {
	ProductID    string
	TariffName   string
	DisplayName  string
	PriceMin     int
	PriceMax     int
	PriceMean    float64
	Eta          int
	TemplateVars map[string]string
}

// ByProductID - упорядочиваем список APIData по возрастанию product id
type ByProductID []APIData

func (a ByProductID) Len() int           { return len(a) }
func (a ByProductID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByProductID) Less(i, j int) bool { return a[i].ProductID < a[j].ProductID }

// APIDataGetter - интерфейс для общения с провайдерами
type APIDataGetter interface {
	GetAPIData(context.Context, *httprequester.Requester, Request) ([]APIData, error)
	APIName() string
}

// ProductsCache - интерфейс для работы с хранилищем продуктов
type ProductsCache interface {
	GetProducts(int) ([]product.Product, error)
	IsOK() bool
}

// DistanceTimeService - интерфейс, который предоставляет время и расстояние проезда от А до Б
type DistanceTimeService interface {
	DistanceTime(ctx context.Context, httpreq *httprequester.Requester, taxiReq Request) (int, int, error)
}

// AddressService - интерфейс, который возвращает адрес в зависимости от координат
type AddressService interface {
	Address(ctx context.Context, lon, lat float64) (*webapi.PointInfo, error)
	AreaNameByLatLon(lat, lon float64) string
}

// Service - струтура для сервиса такси
type Service struct {
	apisMap     map[string]APIDataGetter
	prodCache   ProductsCache
	httpreq     *httprequester.Requester
	distTimeSrv DistanceTimeService
	addrSrv     AddressService
	collector   *collector.Collector
	Logger      *log.StructuredLogger
}

// IsOK - проверяем, работоспособен ли сервис
func (s *Service) IsOK() bool {
	if len(s.apisMap) == 0 {
		return false
	}
	if !s.prodCache.IsOK() {
		return false
	}
	return true
}

func (s *Service) requestOne(ctx context.Context, wg *sync.WaitGroup, taxiReq Request, taxiAPI APIDataGetter, prod product.Product, mutex *sync.Mutex, result *[]serviceRecord) {
	defer wg.Done()
	apiCtx := context.WithValue(ctx, log.CtxKeyAPIName, taxiAPI.APIName())
	start := time.Now()
	taxiData, err := taxiAPI.GetAPIData(apiCtx, s.httpreq, taxiReq)
	elapsed := float64(time.Since(start).Nanoseconds()) / 1000000
	if err != nil {
		s.Logger.ServiceWarningLogEntry(apiCtx, errors.Wrap(err, taxiAPI.APIName()), "request API", taxiAPI.APIName())
		if strings.Contains(err.Error(), ErrInvalidPrice.Error()) {
			s.collector.AddProviderInvalidValueResponse(taxiAPI.APIName(), "price", taxiReq.RegionID)
		}
		if strings.Contains(err.Error(), ErrInvalidTime.Error()) {
			s.collector.AddProviderInvalidValueResponse(taxiAPI.APIName(), "time", taxiReq.RegionID)
		}
	}
	s.collector.AddProviderResponseTime(taxiAPI.APIName(), "all", elapsed)
	s.collector.AddProviderOKResponse(taxiAPI.APIName(), taxiReq.RegionID)
	filteredTaxiData := make([]APIData, 0)
	for _, tData := range taxiData {
		if prod.IsGoodTariff(tData.TariffName) {
			filteredTaxiData = append(filteredTaxiData, tData)
		}
	}
	if len(filteredTaxiData) == 0 && len(taxiData) != 0 {
		s.Logger.ServiceWarningLogEntry(apiCtx, errors.Wrap(ErrNoAPIData, taxiAPI.APIName()), "fitlered all tariffs", taxiAPI.APIName())
		s.collector.AddFilterError(taxiAPI.APIName(), taxiReq.RegionID)
		return
	}
	for _, tData := range filteredTaxiData {
		serviceRecord, err := newServiceRecord(tData, prod)
		if err != nil {
			s.Logger.Warning(errors.Wrap(err, "Error when creating service record, not blocking"))
		}
		mutex.Lock()
		*result = append(*result, serviceRecord)
		mutex.Unlock()
	}
	return
}

func (s *Service) requestProviders(ctx context.Context, taxiReq Request, prods []product.Product) ([]serviceRecord, error) {
	serviceRecords := make([]serviceRecord, 0)
	var wg sync.WaitGroup
	var mutex sync.Mutex
	for _, prod := range prods {
		taxiAPI, ok := s.apisMap[prod.ProviderName]
		if !ok {
			continue
		}
		wg.Add(1)
		go s.requestOne(ctx, &wg, taxiReq, taxiAPI, prod, &mutex, &serviceRecords)
	}
	wg.Wait()
	if len(serviceRecords) == 0 {
		return serviceRecords, errors.Wrap(ErrNoAPIData, "All providers")
	}
	return serviceRecords, nil
}

func (s *Service) getOptimalElse(records []serviceRecord, isOptimalInRegion bool, priceCoeff float64) ([]serviceRecord, []serviceRecord) {
	var optimalCandidateInd = -1

	sort.Sort(byPrice(records))

	if isOptimalInRegion {
		for ind, rec := range records {
			if rec.Operator.IsOptimal && isOptimalInRegion {
				optimalCandidateInd = ind
				break
			}
		}
	}

	if !isOptimalInRegion {
		var suitableCandidates int
		priceLimit := float64(records[0].Price) * priceCoeff
		for k := range records {
			if float64(records[k].Price) <= priceLimit {
				suitableCandidates = k
				continue
			}
			break
		}
		var randInd int
		if suitableCandidates > 0 {
			s := rand.NewSource(time.Now().Unix())
			r := rand.New(s) // initialize local pseudorandom generator
			randInd = r.Intn(suitableCandidates + 1)
		} else {
			randInd = 0
		}
		optimalCandidateInd = randInd
	}

	elses := make([]serviceRecord, 0)

	for ind, rec := range records {
		if ind == optimalCandidateInd {
			continue
		}
		elses = append(elses, rec)
	}

	if optimalCandidateInd == -1 {
		return nil, elses
	}

	return []serviceRecord{records[optimalCandidateInd]}, elses
}

func (s *Service) getServiceRecords(ctx context.Context, wg *sync.WaitGroup, req Request, priceCoeff float64, optimal *[]serviceRecord, elses *[]serviceRecord, errResult *error) {
	defer wg.Done()
	start := time.Now()
	prods, errProds := s.prodCache.GetProducts(req.RegionID)
	if errProds != nil {
		*errResult = errors.Wrap(errProds, "Products not found for region")
		return
	}
	elapsed := float64(time.Since(start).Nanoseconds()) / 1000000
	s.Logger.TimingLogEntry(ctx, elapsed, "Fetch prods")
	start = time.Now()
	var isOptimalInRegion = false
	for _, prod := range prods {
		if prod.IsOptimal {
			isOptimalInRegion = true
			break
		}
	}
	elapsed = float64(time.Since(start).Nanoseconds()) / 1000000
	s.Logger.TimingLogEntry(ctx, elapsed, "Filter prods")
	start = time.Now()
	serviceRecords, errProviders := s.requestProviders(ctx, req, prods)
	if errProviders != nil {
		*errResult = errors.Wrap(errProviders, "Error when request taxiAPIs")
		return
	}
	elapsed = float64(time.Since(start).Nanoseconds()) / 1000000
	s.Logger.TimingLogEntry(ctx, elapsed, "Request providers")

	optimalLocal, elsesLocal := s.getOptimalElse(serviceRecords, isOptimalInRegion, priceCoeff)

	*optimal = optimalLocal
	*elses = elsesLocal
	return
}

func (s *Service) getMeta(ctx context.Context, wg *sync.WaitGroup, taxiReq Request, result *meta, errResult *error) {
	defer wg.Done()
	mosesCtx := context.WithValue(ctx, log.CtxKeyAPIName, "moses")
	dist, tm, errMoses := s.distTimeSrv.DistanceTime(mosesCtx, s.httpreq, taxiReq)
	if errMoses != nil {
		*errResult = errors.Wrap(errMoses, "Error when requesting Moses")
		return
	}
	*result = newMeta(dist, tm)
}

// Response - возвращает ответ с данными от провайдеров такси
func (s *Service) Response(ctx context.Context, req Request, priceCoeff float64) (*Response, error) {
	var errRecords, errMeta error
	var optimal, elses []serviceRecord
	var m meta
	var wg sync.WaitGroup
	response := new(Response)
	wg.Add(2)
	go s.getServiceRecords(ctx, &wg, req, priceCoeff, &optimal, &elses, &errRecords)
	go s.getMeta(ctx, &wg, req, &m, &errMeta)
	wg.Wait()
	if errRecords != nil {
		s.Logger.ServiceWarningLogEntry(ctx, errRecords, "Empty data from providers", "provider")
		s.collector.AddServiceError("providers_empty", req.RegionID)
		return response, errors.Wrap(errRecords, "Empty data from providers")
	}
	if errMeta != nil {
		s.Logger.ServiceWarningLogEntry(ctx, errMeta, "Emty data from Moses", "moses")
		s.collector.AddServiceError("moses", req.RegionID)
	}
	if m.isEmpty() {
		return newResponse(nil, optimal, elses), nil
	}
	return newResponse(&m, optimal, elses), nil
}

// ParseTaxiRequest - парсим запрос к сервису такси
func (s *Service) ParseTaxiRequest(ctx context.Context, r *http.Request) (Request, error) {
	taxiReqID := middleware.GetReqID(r.Context())
	var taxiReq Request
	taxiReq.ReqID = taxiReqID
	content, errRead := ioutil.ReadAll(r.Body)
	if errRead != nil {
		s.collector.AddServiceError("invalid_request", taxiReq.RegionID)
		return taxiReq, errors.Wrap(errRead, "Cannot load request body")
	}
	r.Body.Close()
	if err := json.NewDecoder(bytes.NewReader(content)).Decode(&taxiReq); err != nil {
		s.collector.AddServiceError("invalid_request", taxiReq.RegionID)
		return taxiReq, errors.Wrap(err, "Cannot parse request json")
	}

	if taxiReq.RegionID == 0 || taxiReq.Point1.IsEmpty() || taxiReq.Point2.IsEmpty() {
		return taxiReq, ErrTaxiReqEmpty
	}
	return taxiReq, nil
}

// EvaluateTaxiRequest - парсим запрос к сервису такси
func (s *Service) EvaluateTaxiRequest(ctx context.Context, taxiReq *Request) error {
	start := time.Now()
	var wg sync.WaitGroup
	var addrPoint1, addrPoint2 *webapi.PointInfo
	var errAddrPoint1, errAddrPoint2 error
	wg.Add(2)
	go func() {
		defer wg.Done()
		addrPoint1, errAddrPoint1 = s.addrSrv.Address(ctx, taxiReq.Point1.Lat, taxiReq.Point1.Lon)
	}()
	go func() {
		defer wg.Done()
		addrPoint2, errAddrPoint2 = s.addrSrv.Address(ctx, taxiReq.Point2.Lat, taxiReq.Point2.Lon)
	}()
	wg.Wait()
	if errAddrPoint1 != nil {
		s.collector.AddServiceError("webapi_point", taxiReq.RegionID)
		return errors.Wrap(errAddrPoint1, "Cannot evaluate address for point1")
	}
	if errAddrPoint2 != nil {
		s.collector.AddServiceError("webapi_point", taxiReq.RegionID)
		return errors.Wrap(errAddrPoint2, "Cannot evaluate address for point2")
	}
	elapsed := float64(time.Since(start).Nanoseconds()) / 1000000
	s.Logger.TimingLogEntry(ctx, elapsed, "Evaluate request: request webAPI")
	start = time.Now()
	taxiReq.Point1.AddGeoInfo(addrPoint1)
	taxiReq.Point2.AddGeoInfo(addrPoint2)

	taxiReq.Point1.AddArea(s.addrSrv.AreaNameByLatLon(taxiReq.Point1.Lat, taxiReq.Point1.Lon))
	taxiReq.Point2.AddArea(s.addrSrv.AreaNameByLatLon(taxiReq.Point2.Lat, taxiReq.Point2.Lon))
	elapsed = float64(time.Since(start).Nanoseconds()) / 1000000
	s.Logger.TimingLogEntry(ctx, elapsed, "Evaluate request: Parse area")
	return nil
}
