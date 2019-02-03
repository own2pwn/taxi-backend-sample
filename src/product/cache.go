package product

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/collector"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/log"
)

// ErrNoProducts - продукт в базе не найден
var ErrNoProducts = errors.New("Products not found in Cache")

type regionToProducts map[int][]Product

// Storage - interface for produvts data storage
type Storage interface {
	GetAllProducts() (regionToProducts, error)
}

// Cache - cache Storage
type Cache struct {
	storage Storage
	cache   regionToProducts
	cacheMu *sync.Mutex
}

// NewCache - create new cache
func NewCache(storage Storage, coll *collector.Collector, logger log.Logger) (*Cache, error) {
	cache, err := storage.GetAllProducts()
	if err != nil {
		return new(Cache), err
	}
	return &Cache{
		storage: storage,
		cache:   cache,
		cacheMu: &sync.Mutex{},
	}, nil
}

// GetProducts - возвращает все продукты для региона
func (c Cache) GetProducts(regionID int) ([]Product, error) {
	c.cacheMu.Lock()
	products, ok := c.cache[regionID]
	c.cacheMu.Unlock()
	if ok {
		return products, nil
	}
	return nil, errors.Wrapf(ErrNoProducts, "region: %v", regionID)
}

// IsOK - проверяем, есть ли объекты в кэше
func (c Cache) IsOK() bool {
	c.cacheMu.Lock()
	length := len(c.cache)
	c.cacheMu.Unlock()
	return length != 0
}

// Reload - перезагружаем кэш
func (c *Cache) Reload() error {
	updatedCache, err := c.storage.GetAllProducts()
	if err != nil {
		return err
	}
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.cache = updatedCache
	return nil
}
