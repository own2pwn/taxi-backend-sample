package service

import (
	"context"
	"encoding/json"
	"net/url"
	"path"

	"github.com/pkg/errors"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/httprequester"
)

var (
	mosesOutput    = "simple"
	mosesReqType   = "jam"
	mosesPointType = "stop"
	// ErrMosesEmptyResult - от Моисея пришел пустой ответ
	ErrMosesEmptyResult = errors.New("Moses returned empty result")
	// ErrMosesInvalidResult - от Моисея пришел ответ c 0 0
	ErrMosesInvalidResult = errors.New("Moses returned invalid result")
	mosesURL              = "http://routing.2gis.com/carrouting/4.0.0/"
)

type mosesPoint struct {
	Type string  `json:"type"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

type mosesReqData struct {
	Output string       `json:"output"`
	Type   string       `json:"type"`
	Points []mosesPoint `json:"points"`
}

type mosesRes struct {
	Type     string
	Duration int
	Length   int
}

type mosesAnswer struct {
	Query  mosesReqData
	Result []mosesRes
	Type   string
}

// RegionsInfo - интерфейс к сущности, у которой можно запрашивать имя региона по коду региона
type RegionsInfo interface {
	GetRegionNameByID(int) (string, error)
}

// MosesService - структура для сервиса, который делает запросы к Моисею
type MosesService struct {
	mosesURL    *url.URL
	regionsInfo RegionsInfo
}

// NewMosesService - создаем новый сервис для запросов к Моисею
func NewMosesService(regionsInfo RegionsInfo) (*MosesService, error) {
	mosesURL, errParse := url.Parse(mosesURL)
	if errParse != nil {
		return nil, errors.Wrapf(errParse, "Moses: Cannot parse Moses URL: %v", mosesURL)
	}
	return &MosesService{
		mosesURL,
		regionsInfo,
	}, nil
}

// DistanceTime - возвращает дистанцию и время проезда от точки А в Б
func (m *MosesService) DistanceTime(ctx context.Context, httpreq *httprequester.Requester, taxiReq Request) (int, int, error) {
	regionName, errRegionName := m.regionsInfo.GetRegionNameByID(taxiReq.RegionID)
	if errRegionName != nil {
		return 0, 0, errors.Wrap(errRegionName, "Moses: Cannot get region name")
	}
	data := mosesReqData{
		Output: mosesOutput,
		Type:   mosesReqType,
		Points: []mosesPoint{
			{
				Type: mosesPointType,
				X:    taxiReq.Point1.Lon,
				Y:    taxiReq.Point1.Lat,
			},
			{
				Type: mosesPointType,
				X:    taxiReq.Point2.Lon,
				Y:    taxiReq.Point2.Lat,
			},
		},
	}
	bData, errData := json.Marshal(data)
	if errData != nil {
		return 0, 0, errors.Wrap(errData, "Moses: Cannot marshall data for Moses request")
	}
	headers := []httprequester.Dict{
		{
			Key:   "X-Internal-Service",
			Value: "TAXA",
		},
		{
			Key:   "Content-Type",
			Value: "application/json",
		},
	}
	var p mosesAnswer
	modifiedURL := *m.mosesURL
	modifiedURL.Path = path.Join(m.mosesURL.Path, regionName)
	err := httpreq.Post(ctx, modifiedURL.String(), headers, bData, &p)
	if err != nil {
		return 0, 0, errors.Wrap(err, "Moses: Cannot request Moses")
	}
	if len(p.Result) == 0 {
		return 0, 0, ErrMosesEmptyResult
	}
	if p.Result[0].Length < 0 || p.Result[0].Duration < 0 {
		return 0, 0, errors.Wrapf(ErrMosesInvalidResult, "Length %v, duration %v", p.Result[0].Length <= 0, p.Result[0].Duration)
	}
	return p.Result[0].Length, p.Result[0].Duration / 60, nil
}
