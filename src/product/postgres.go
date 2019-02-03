package product

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/database"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/log"
)

var (
	// ErrImageNotFound - не смогли найти ссылку на изображение
	ErrImageNotFound = errors.New("Cannot find image in map")
	// ErrProduct - не смогли создать продукт из записи базы
	ErrProduct = errors.New("Error when creating product from record")
	// ErrRecordEmpty - обязательное поле в записи базы пустое
	ErrRecordEmpty = errors.New("DB record: required feild is empty")
)

var imgAlternativeNames = map[string]string{
	"rutaxi":    "rutaxi_1",
	"citymobil": "citymobil_1",
}

const allRecQuery = "SELECT id, region_id, name, title, short_title, site_caption, site_value, app_url, phone_caption, phone_value, android_app_url, android_app_id, ios_app_url, ios_app_id, api_org_id, api_id, api_data, rating, avg_eta, is_active, currency_code, handler, is_optimal FROM provider WHERE is_active=TRUE and handler != ''"

func imagePath(name string) string {
	if altName, ok := imgAlternativeNames[name]; ok {
		name = altName
	}
	return fmt.Sprintf("https://disk.2gis.com/taksa-providers/provider_%v.png", name)
}

type record struct {
	ID            int      `db:"id"`
	RegionID      int      `db:"region_id"`
	Name          string   `db:"name"`
	Title         string   `db:"title"`
	ShortTitle    *string  `db:"short_titile"`
	SiteCaption   *string  `db:"site_caption"`
	SiteValue     *string  `db:"site_value"`
	AppURL        *string  `db:"app_url"`
	PhoneCaption  *string  `db:"phone_caption"`
	PhoneValue    *string  `db:"phone_value"`
	AndroidAppURL *string  `db:"android_app_url"`
	AndroidAppID  *string  `db:"android_app_id"`
	IosAppURL     *string  `db:"ios_app_url"`
	IosAppID      *string  `db:"ios_app_id"`
	APIOrgID      int64    `db:"api_org_id"`
	APIID         int64    `db:"api_id"`
	APIData       *string  `db:"api_data"`
	Rating        *float32 `db:"rating"`
	AvgEta        *int     `db:"avg_eta"`
	IsActive      *bool    `db:"is_active"`
	CurrencyCode  *string  `db:"currency_code"`
	Handler       string   `db:"handler"`
	IsOptimal     bool     `db:"is_optimal"`
}

func (r *record) name() string {
	nameParts := strings.Split(r.Name, ":")
	if len(nameParts) < 1 {
		return ""
	}
	return nameParts[0]
}

func (r *record) tariff() string {
	namePatrs := strings.Split(r.Name, ":")
	if len(namePatrs) < 2 || namePatrs[1] == namePatrs[0] {
		return ""
	}
	return namePatrs[1]
}

func (r record) IsEmpty() bool {
	//if r.RegionID == 0 || r.APIOrgID == 0 || r.Handler == "" || r.CurrencyCode == nil {
	if r.RegionID == 0 || r.Handler == "" || r.CurrencyCode == nil {
		return true
	}
	return false
}

// NewPostgresRep - кэш постгрес базы
func NewPostgresRep(db *database.Database, logger log.Logger) Storage {
	return postgresRep{
		db:     db,
		logger: logger,
	}
}

type postgresRep struct {
	db     *database.Database
	logger log.Logger
}

func (pg postgresRep) GetAllProducts() (regionToProducts, error) {
	records, errRec := pg.getAllRecords()
	if errRec != nil {
		return nil, errors.Wrap(errRec, "Getting all products from Postgres:")
	}
	regionMap := make(map[int][]record)
	for _, rec := range records {
		regID := rec.RegionID
		regionMap[regID] = append(regionMap[regID], rec)
	}

	regionProductMap := make(regionToProducts)
	for regID, recs := range regionMap {
		prods, err := pg.recordsToProducts(recs)
		if err != nil {
			pg.logger.Warning(errors.Wrap(err, "Cannot make region products"))
		}
		if len(prods) != 0 {
			regionProductMap[regID] = prods
		}
	}
	return regionProductMap, nil
}

func (pg postgresRep) recordsToProducts(records []record) ([]Product, error) {
	products := make([]Product, 0)
	productsMap := make(map[string]Product)
	for _, rec := range records {
		mapKey := rec.name()
		product, ok := productsMap[mapKey]
		if !ok {
			prod, err := pg.recordToProduct(rec)
			if err != nil {
				return products, errors.Wrapf(err, "Cannot make product for record %#v", rec)
			}
			productsMap[mapKey] = prod
			continue
		}
		tariff := rec.tariff()
		product.AddTariff(tariff)
		productsMap[mapKey] = product
	}

	for _, prod := range productsMap {
		products = append(products, prod)
	}
	return products, nil
}

func (pg postgresRep) getAllRecords() ([]record, error) {
	rows, err := pg.db.Query(allRecQuery)
	if err != nil {
		return nil, errors.Wrap(err, "Prostgres DB: could not select all products")
	}
	defer rows.Close()
	records := make([]record, 0)
	for rows.Next() {
		var p record
		errScan := rows.Scan(&p.ID, &p.RegionID, &p.Name, &p.Title, &p.ShortTitle, &p.SiteCaption, &p.SiteValue, &p.AppURL, &p.PhoneCaption, &p.PhoneValue, &p.AndroidAppURL, &p.AndroidAppID, &p.IosAppURL, &p.IosAppID, &p.APIOrgID, &p.APIID, &p.APIData, &p.Rating, &p.AvgEta, &p.IsActive, &p.CurrencyCode, &p.Handler, &p.IsOptimal)
		if errScan != nil {
			return nil, errors.Wrap(errScan, "Prostgres DB: could not scan provider")
		}
		records = append(records, p)
	}
	for _, rec := range records {
		if rec.IsEmpty() {
			return nil, errors.Wrapf(ErrRecordEmpty, "Record: %#v", rec)
		}
	}
	return records, nil
}

func (pg postgresRep) recordToProduct(rec record) (Product, error) {
	ImgURL, errImg := pg.getImage(rec.name())
	if errImg != nil {
		pg.logger.Warningf("No image: %v", errImg)
	}
	tariffs := make([]string, 0)
	tariff := rec.tariff()
	if tariff != "" {
		tariffs = append(tariffs, tariff)
	}
	prod := Product{
		ID:             rec.ID,
		RegionID:       rec.RegionID,
		Name:           rec.name(),
		Tariffs:        tariffs,
		ShortTitle:     rec.ShortTitle,
		SiteCaption:    rec.SiteCaption,
		SiteValue:      rec.SiteValue,
		AppURLTemplate: rec.AppURL,
		PhoneCaption:   rec.PhoneCaption,
		PhoneValue:     rec.PhoneValue,
		AndroidAppURL:  rec.AndroidAppURL,
		AndroidAppID:   rec.AndroidAppID,
		IosAppURL:      rec.IosAppURL,
		IosAppID:       rec.IosAppID,
		APIID:          rec.APIID,
		APIOrgID:       rec.APIOrgID,
		APIData:        rec.APIData,
		Rating:         rec.Rating,
		AvgEta:         rec.AvgEta,
		CurrencyCode:   rec.CurrencyCode,
		ProviderName:   rec.Handler,
		ImageURL:       ImgURL,
		IsOptimal:      rec.IsOptimal,
	}
	if prod.IsEmpty() {
		return prod, errors.Wrap(ErrProduct, "No data to create product")
	}
	return prod, nil
}

func (pg postgresRep) getImage(name string) (*url.URL, error) {
	imagePath := imagePath(name)
	imgURL, err := url.Parse(imagePath)
	if err != nil {
		return imgURL, errors.Wrapf(err, "Cannot parse Image URL: %v", imagePath)
	}
	return imgURL, nil
}
