package pointresolver

import (
	"strconv"

	"strings"

	"github.com/golang/geo/s2"
	"github.com/pkg/errors"
)

var (
	// ErrLonLat - не удалось распарсить координаты из файла areas.json
	ErrLonLat = errors.New("Cannot parse lon and lat")
)

// area - структура, которая хранит в себе область: название и геометрию
type area struct {
	Name        string   `json:"name"`
	GeometryStr []string `json:"geometry"`
	Loops       []*s2.Loop
}

// containsPoint - проверяем, входит ли точка в область
func (a *area) containsPoint(point *s2.Point) bool {
	for _, loop := range a.Loops {
		if loop.ContainsPoint(*point) {
			return true
		}
	}
	return false
}

// areas - структуры, в которой хранится список всех областей из areas.json
type areas struct {
	areas []area
}

// areaNameByLatLon - возвращает имя области, в которую входит точка.
// Если точка не входит ни в какую область, то возвращается пустая срока
func (a *areas) areaNameByLatLon(lat, lon float64) string {
	point := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lon))
	for _, area := range a.areas {
		if area.containsPoint(&point) {
			return area.Name
		}
	}
	return ""
}

// newAreas - создаем структуру, в которую складываются все области из areas.json
func newAreas() (*areas, error) {
	a := areasData
	for i, area := range a.areas {
		loops := make([]*s2.Loop, 0)
		for _, coordSet := range area.GeometryStr {
			loop, err := parseGeometry(coordSet)
			if err != nil {
				return &a, err
			}
			loops = append(loops, loop)
		}
		a.areas[i].Loops = loops
	}
	return &a, nil
}

func parseGeometry(coordStr string) (*s2.Loop, error) {
	coords := strings.Split(string(coordStr), ",")
	points := make([]s2.Point, 0)
	for _, coord := range coords {
		coord = strings.TrimSpace(coord)
		lonLat := strings.Split(coord, " ")
		if len(lonLat) < 2 {
			return nil, errors.Wrapf(ErrLonLat, "Cannot split coord pair: %v", coord)
		}
		lon, errLon := strconv.ParseFloat(lonLat[0], 64)
		if errLon != nil {
			return nil, errors.Wrapf(ErrLonLat, "Cannot convert string to Lon float64: %v", lonLat[0])
		}
		lat, errLat := strconv.ParseFloat(lonLat[1], 64)
		if errLat != nil {
			return nil, errors.Wrapf(ErrLonLat, "Cannot convert string to Lat float64: %v", lonLat[1])
		}
		point := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lon))
		points = append(points, point)
	}
	loop := s2.LoopFromPoints(points)
	return loop, nil
}
