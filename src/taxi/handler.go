package taxi

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/log"
	"github.com/nburunova/taxi-backend-sample/src/taxi/service"
)

// RegionPriceCoeff - тип для хранения данных о коэффициенте цен в проектах
type RegionPriceCoeff map[int]float64

// GetByRegionOrElse - получить коэффициент в зависимости от кода проекта, если для этого региона есть свой коэффициент
func (r RegionPriceCoeff) GetByRegionOrElse(regID int, elseOption float64) float64 {
	if regionCoeff, ok := r[regID]; ok {
		return regionCoeff
	}
	return elseOption
}

// Service - интерфейс сервиса такси
type Service interface {
	ParseTaxiRequest(ctx context.Context, r *http.Request) (service.Request, error)
	Response(ctx context.Context, req service.Request, priceCoeff float64) (*service.Response, error)
	EvaluateTaxiRequest(ctx context.Context, taxiReq *service.Request) error
}

type handlerFunc func(service Service, priceCoeff float64, regPriceCoeff RegionPriceCoeff, waitTime time.Duration, w http.ResponseWriter, r *http.Request)

// SomeHandler - обертка над хэндлером запроса данных такси
func SomeHandler(service Service, basicPriceCoeff float64, regPriceCoeff RegionPriceCoeff, waitTime time.Duration, handler handlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		handler(service, basicPriceCoeff, regPriceCoeff, waitTime, w, r)
	}
	return http.HandlerFunc(fn)
}

// Handler - хендлер запроса данных такси
func Handler(service Service, basicPriceCoeff float64, regPriceCoeff RegionPriceCoeff, waitTime time.Duration, w http.ResponseWriter, r *http.Request) {
	ctxTaxi, cancelTaxi := context.WithTimeout(r.Context(), waitTime)
	defer cancelTaxi()
	taxiReq, taxiReqParseErr := service.ParseTaxiRequest(ctxTaxi, r)
	if taxiReqParseErr != nil {
		render.Render(w, r, errInvalidRequest(taxiReqParseErr))
		return
	}
	ctxTaxi = context.WithValue(ctxTaxi, log.CtxKeyRegionID, taxiReq.RegionID)
	errEval := service.EvaluateTaxiRequest(ctxTaxi, &taxiReq)
	if errEval != nil {
		render.Render(w, r, errServerError(errEval))
		return
	}
	coeff := regPriceCoeff.GetByRegionOrElse(taxiReq.RegionID, basicPriceCoeff)
	response, errResponse := service.Response(ctxTaxi, taxiReq, coeff)
	if errResponse != nil {
		render.Render(w, r, errServerError(errResponse))
		return
	}
	render.JSON(w, r, response)
}
