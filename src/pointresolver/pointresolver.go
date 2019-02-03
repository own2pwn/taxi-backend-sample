package pointresolver

import (
	"context"

	"github.com/pkg/errors"
	"github.com/nburunova/taxi-backend-sample/src/webapi"
)

//PointResolver - структура для получения дополнительной иформации о гео точке
type PointResolver struct {
	webAPIcl *webapi.Client
	ars      *areas
}

// NewPointResolver - создает объект типа PointResolver, который отдает строку адреса и область точки (lat, lon)
func NewPointResolver(wAPIcl *webapi.Client) (*PointResolver, error) {
	ar, errAr := newAreas()
	if errAr != nil {
		return nil, errors.Wrap(errAr, "PointResolver: cannot init areas")
	}
	return &PointResolver{
		wAPIcl,
		ar,
	}, nil
}

// Address - возвращает ближайший адрес к точке в виде строки
func (pr *PointResolver) Address(ctx context.Context, lat, lon float64) (*webapi.PointInfo, error) {
	return pr.webAPIcl.GetPointInfo(ctx, lat, lon)
}

// AreaNameByLatLon - возвращает имя области, в которую входит точка.
// Если точка не входит ни в какую область, то возвращается пустая срока
func (pr *PointResolver) AreaNameByLatLon(lat, lon float64) string {
	return pr.ars.areaNameByLatLon(lat, lon)
}
