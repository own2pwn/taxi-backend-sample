package service

import (
	"strconv"

	"github.com/nburunova/taxi-backend-sample/src/webapi"
)

// Point - струтура для точки
type Point struct {
	Lat     float64 `json:"lat"`
	LatStr  string
	Lon     float64 `json:"lon"`
	LonStr  string
	Address string
	Area    string
}

func (p *Point) stringfy() {
	p.LatStr = strconv.FormatFloat(p.Lat, 'f', -1, 64)
	p.LonStr = strconv.FormatFloat(p.Lon, 'f', -1, 64)
}

// IsEmpty - проверяем, есть ли значения lat & lon у запроса
func (p *Point) IsEmpty() bool {
	return p.Lat == 0 || p.Lon == 0
}

// AddGeoInfo - обновляем информацию о точке
func (p *Point) AddGeoInfo(pInfo *webapi.PointInfo) {
	p.Address = pInfo.Address
	p.Lat = pInfo.Lat
	p.Lon = pInfo.Lon
	p.stringfy()
}

// AddArea - обновляем информацию об области, в которой находится точка
func (p *Point) AddArea(area string) {
	p.Area = area
}

// Request - струтура, описывающая запрос к сервису такси
type Request struct {
	ReqID    string
	RegionID int   `json:"region_id" json:"req_region"`
	Point1   Point `json:"point1" json:"point1"`
	Point2   Point `json:"point2" json:"point2"`
	OnlyAPI  bool  `json:"only_api"`
}
