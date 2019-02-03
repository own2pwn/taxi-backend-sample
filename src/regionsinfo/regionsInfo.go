package regionsinfo

import (
	"strconv"
	"sync"

	"github.com/nburunova/taxi-backend-sample/src/webapi"
	"github.com/pkg/errors"
)

var (
	// ErrRegionNameNotFound - не нашли имя для айди региона
	ErrRegionNameNotFound = errors.New("Region name not found by ID")
)

// RegionsInfo - структура, в которой хранятся соотвествия "код региона" - "имя региона"
type RegionsInfo struct {
	regionCodeToName map[int]string
	updateMu         *sync.Mutex
}

// NewRegionsInfo - создаем новую структуру для хранителя кодов и имен регионов
func NewRegionsInfo() *RegionsInfo {
	return &RegionsInfo{
		map[int]string{},
		&sync.Mutex{},
	}
}

// GetRegionNameByID - имя региона по коду региона
func (r *RegionsInfo) GetRegionNameByID(regionID int) (string, error) {
	r.updateMu.Lock()
	defer r.updateMu.Unlock()
	if regionName, ok := r.regionCodeToName[regionID]; ok {
		return regionName, nil
	}
	return "", errors.Wrap(ErrRegionNameNotFound, strconv.Itoa(regionID))
}

// Load - обновить список регионов новыми данными
func (r *RegionsInfo) Load(regionsInfo []webapi.RegionInfo) error {
	newRegionToName := map[int]string{}
	for _, regionData := range regionsInfo {
		newRegionToName[regionData.ID] = regionData.Name
	}
	r.updateMu.Lock()
	defer r.updateMu.Unlock()
	r.regionCodeToName = newRegionToName
	return nil
}
