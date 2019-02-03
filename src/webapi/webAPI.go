package webapi

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/httprequester"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/log"
)

var (
	// ErrAddressNotFound - не нашли адрес по координатам
	ErrAddressNotFound = errors.New("Address not found")
	// ErrNotPoint - не смогли распарсить точки из ответа ВебАПИ
	ErrNotPoint = errors.New("Cannot parse POINT from centroid string")
	// ErrMeta - мета код в ответе ВебАПИ не 200
	ErrMeta = errors.New("Meta code != 200")
	// ErrEmptyResult - из ВебАПИ вернулся пустой список геообъектов
	ErrEmptyResult = errors.New("Result list is empty")

	geoParams = []httprequester.Dict{
		{
			Key:   "key",
			Value: "ruvgco0172",
		},
		{
			Key:   "type",
			Value: "building",
		},
		{
			Key:   "fields",
			Value: "items.context,items.geometry.centroid",
		},
		{
			Key:   "page",
			Value: "1",
		},
		{
			Key:   "page_size",
			Value: "1",
		},
		{
			Key:   "radius",
			Value: "250",
		},
	}
	regionListParams = []httprequester.Dict{
		{
			Key:   "key",
			Value: "navidev",
		},
		{
			Key:   "fields",
			Value: "items.code",
		},
	}
	webAPIURL        = "http://catalog.api.2gis.ru"
	geoMethod        = "/2.0/geo/search"
	regionListMethod = "/2.0/region/list"
	reLeadCloseWhtsp = regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	reInsideWhtsp    = regexp.MustCompile(`[\s\p{Zs}]{2,}`)
)

type meta struct {
	Code       int    `json:"code"`
	APIVersion string `json:"api_version"`
	IssueDate  string `json:"issue_date"`
}

type geo struct {
	PurposeName string `json:"purpose_name"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Context     struct {
		Dist     float64 `json:"dist"`
		Distance int     `json:"distance"`
	} `json:"context"`
	ID          string `json:"id"`
	AddressName string `json:"address_name"`
	Type        string `json:"type"`
	Geometry    struct {
		Centroid string `json:"centroid"`
	} `json:"geometry"`
}

type webAPIGeoAnwser struct {
	Meta   meta `json:"meta"`
	Result struct {
		Total int   `json:"total"`
		Items []geo `json:"items"`
	} `json:"result"`
}

// PointInfo - данные о точке - адрес, широта и долгота объекта
type PointInfo struct {
	Address  string
	Lat, Lon float64
}

type region struct {
	Name string `json:"name"`
	Code string `json:"code"`
	ID   string `json:"id"`
	Type string `json:"region"`
}

type webAPIRegionAnswer struct {
	Meta   meta `json:"meta"`
	Result struct {
		Total int      `json:"total"`
		Items []region `json:"items"`
	} `json:"result"`
}

// RegionInfo - данные о регионе - имя и код
type RegionInfo struct {
	Name string
	ID   int
}

// Client - структура сервиса для запросов к WebAPI
type Client struct {
	httpreq *httprequester.Requester
}

// NewClient - создаем структуру для сервиса запросов к WebAPI
func NewClient(httpreq *httprequester.Requester) (*Client, error) {
	return &Client{
		httpreq,
	}, nil
}

// GetPointInfo - возвращает ближайший геообъект к точке - адрес в виде строки и координаты
func (c *Client) GetPointInfo(ctx context.Context, lat, lon float64) (*PointInfo, error) {
	webAPICtx := context.WithValue(ctx, log.CtxKeyAPIName, "webAPI")
	params := []httprequester.Dict{
		{
			Key:   "point",
			Value: fmt.Sprintf("%v,%v", lon, lat),
		},
	}
	params = append(params, geoParams...)
	pointInfo := new(PointInfo)
	answer := webAPIGeoAnwser{}
	url := fmt.Sprintf("%v%v", webAPIURL, geoMethod)

	err := c.httpreq.Get(webAPICtx, url, nil, params, &answer)
	if err != nil {
		return pointInfo, errors.Wrap(err, "webAPI: Cannot request webAPI geo point info")

	}
	if len(answer.Result.Items) == 0 {
		return pointInfo, ErrEmptyResult
	}
	if answer.Result.Items[0].AddressName == "" && answer.Result.Items[0].Name == "" {
		return pointInfo, ErrAddressNotFound
	}
	var errCentroid error
	pointInfo.Lat, pointInfo.Lon, errCentroid = pointToLatLon(answer.Result.Items[0].Geometry.Centroid)
	if errCentroid != nil {
		return pointInfo, errors.Wrap(errCentroid, "webAPI: Cannot parse centroid")
	}
	if answer.Result.Items[0].AddressName != "" {
		pointInfo.Address = answer.Result.Items[0].AddressName
		return pointInfo, nil
	}
	pointInfo.Address = answer.Result.Items[0].Name
	return pointInfo, nil
}

func pointToLatLon(centroid string) (float64, float64, error) {
	centroid = reLeadCloseWhtsp.ReplaceAllString(centroid, "")
	centroid = reInsideWhtsp.ReplaceAllString(centroid, " ")
	centroid = strings.ToLower(centroid)

	if !(strings.HasPrefix(centroid, "point")) {
		return 0, 0, errors.Wrap(ErrNotPoint, "POINT keyword not found")
	}
	startInd, endInd := strings.Index(centroid, "("), strings.Index(centroid, ")")
	if startInd == -1 || endInd == -1 {
		return 0, 0, errors.Wrap(ErrNotPoint, "( and/or ) not found")
	}
	point := strings.TrimSpace(centroid[startInd+1 : endInd])
	coords := strings.Split(point, " ")
	if len(coords) < 2 {
		return 0, 0, errors.Wrap(ErrNotPoint, "Wrong coord format")
	}
	lonStr, latStr := coords[0], coords[1]
	lat, errLat := strconv.ParseFloat(latStr, 64)
	if errLat != nil {
		return 0, 0, errors.Wrap(errLat, "Cannot parse lat from centroid")
	}
	lon, errLon := strconv.ParseFloat(lonStr, 64)
	if errLon != nil {
		return 0, 0, errors.Wrap(errLon, "Cannot parse lon from centroid")
	}
	return lat, lon, nil
}

//GetRegionsList - отдает список регионов в 2гис
func (c *Client) GetRegionsList(ctx context.Context) ([]RegionInfo, error) {
	webAPICtx, webAPICtxCancel := context.WithTimeout(ctx, 5*time.Second)
	defer webAPICtxCancel()
	webAPICtx = context.WithValue(webAPICtx, log.CtxKeyAPIName, "webAPI")
	answer := webAPIRegionAnswer{}
	url := fmt.Sprintf("%v%v", webAPIURL, regionListMethod)
	err := c.httpreq.Get(webAPICtx, url, nil, regionListParams, &answer)
	if err != nil {
		return nil, errors.Wrap(err, "webAPI: Cannot request webAPI regions list")
	}
	if answer.Meta.Code != 200 {
		return nil, errors.Wrapf(ErrMeta, string(answer.Meta.Code))
	}
	if len(answer.Result.Items) == 0 {
		return nil, ErrEmptyResult
	}
	regionInfo := make([]RegionInfo, len(answer.Result.Items))
	for i, regRawInfo := range answer.Result.Items {
		regID, err := strconv.Atoi(regRawInfo.ID)
		if err != nil {
			return nil, errors.Wrap(err, "webAPI: cannot convert region ID to int")
		}
		regionInfo[i] = RegionInfo{
			Name: regRawInfo.Code,
			ID:   regID,
		}
	}
	return regionInfo, nil
}
