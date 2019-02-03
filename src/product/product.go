package product

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/errorswrapper"
)

const (
	textColor       = "#000"
	backgroundColor = "#A0C305"
)

// ErrNoURL - не задан URL
var ErrNoURL = errors.New("No URL")

// Operator - стуртура, которая используется в ответе сервиса
type Operator struct {
	BranchID        *string    `json:"branch_id"`
	URL             *string    `json:"url"`
	Image           *string    `json:"image"`
	Site            *site      `json:"site"`
	BackgroundColor string     `json:"background_color"`
	ShortTitle      *string    `json:"short_title"`
	StoreURLs       *storeURLs `json:"store_urls"`
	ID              *int       `json:"id"`
	TextColor       string     `json:"text_color"`
	Title           *string    `json:"title"`
	OrgID           *string    `json:"org_id"`
	Phone           *phone     `json:"phone"`
	IsOptimal       bool       `json:"-"`
}

// Product - структура данных для продукта
type Product struct {
	ID             int
	RegionID       int
	Name           string
	Tariffs        []string
	Title          string
	ShortTitle     *string
	SiteCaption    *string
	SiteValue      *string
	AppURLTemplate *string
	PhoneCaption   *string
	PhoneValue     *string
	AndroidAppURL  *string
	AndroidAppID   *string
	IosAppURL      *string
	IosAppID       *string
	APIOrgID       int64
	APIID          int64
	APIData        *string
	Rating         *float32
	AvgEta         *int
	ProviderName   string
	CurrencyCode   *string
	ImageURL       *url.URL
	IsOptimal      bool
}

// IsEmpty - проверяет, заполнилась ли структура данными из базы или дефолтными значениями
func (p *Product) IsEmpty() bool {
	return p.Name == ""
}

// AddTariff - добавляет тариф к продукту
func (p *Product) AddTariff(tariff string) {
	if tariff != "" {
		p.Tariffs = append(p.Tariffs, tariff)
	}
}

// GetOperator - возвращает объект Operator для формирования ответа сервиса
func (p *Product) GetOperator(displayName string, urlReplace map[string]string) (Operator, error) {
	var apiID, apiOrgID string
	var deeplinkStr, imageURLStr *string
	var errorsOperator []error
	apiID = strconv.FormatInt(p.APIID, 10)
	apiOrgID = strconv.FormatInt(p.APIOrgID, 10)
	if p.AppURLTemplate != nil && *p.AppURLTemplate != "" {
		deeplink, deeplinkErr := makeLink(*p.AppURLTemplate, urlReplace)
		if deeplinkErr != nil {
			errorsOperator = append(errorsOperator, deeplinkErr)
		} else {
			linkStr := fmt.Sprintf("%v://%v%v?%v", deeplink.Scheme, deeplink.Host, deeplink.Path, deeplink.RawQuery)
			deeplinkStr = &linkStr
		}
	}
	if p.ImageURL != nil {
		s := p.ImageURL.String()
		imageURLStr = &s
	}
	storeURLs, errStoreURLs := newStoreURLs(p.IosAppID, p.IosAppURL, p.AndroidAppID, p.AndroidAppURL, urlReplace)
	if errStoreURLs != nil {
		errorsOperator = append(errorsOperator, errStoreURLs)
	}

	if displayName == "" {
		displayName = p.Title
	}

	return Operator{
		IsOptimal:       p.IsOptimal,
		URL:             deeplinkStr,
		Site:            newSite(p.SiteValue, p.SiteCaption),
		StoreURLs:       storeURLs,
		Image:           imageURLStr,
		BranchID:        &apiID,
		OrgID:           &apiOrgID,
		BackgroundColor: backgroundColor,
		TextColor:       textColor,
		ShortTitle:      p.ShortTitle,
		ID:              &p.ID,
		Title:           &displayName,
		Phone:           newPhone(p.PhoneValue, p.PhoneCaption),
	}, errorswrapper.WrapErrorSlice(errorsOperator)
}

// IsGoodTariff - проверяет, находится ли тариф в белом списке
func (p *Product) IsGoodTariff(tariff string) bool {
	// если в списке тарифов одно значение -
	if len(p.Tariffs) == 0 {
		return true
	}
	for _, pTariff := range p.Tariffs {
		if pTariff == tariff {
			return true
		}
	}
	return false
}

func makeLink(linkBase string, urlReplaces map[string]string) (*url.URL, error) {
	replacedStr := replace(urlReplaces, linkBase)
	link, err := url.Parse(replacedStr)
	if err != nil {
		return nil, err
	}
	params := link.Query()
	link.RawQuery = params.Encode()
	return link, nil
}

func replace(urlReplaces map[string]string, baseStr string) string {
	modifiedStr := baseStr
	for oldStr, newStr := range urlReplaces {
		modifiedStr = strings.Replace(modifiedStr, oldStr, newStr, -1)
	}
	return modifiedStr
}

type site struct {
	Value *string `json:"value"`
	Text  *string `json:"text"`
}

func newSite(value *string, text *string) *site {
	if value != nil && *value != "" {
		st := site{
			Value: value,
			Text:  text,
		}
		return &st
	}
	return nil
}

type storeURL struct {
	ID  string `json:"id,omitempty"`
	URL string `json:"url,omitempty"`
}

func newStoreURL(ID, someURLStr *string, urlReplaces map[string]string) (*storeURL, error) {
	if someURLStr != nil && *someURLStr != "" {
		someURL, err := makeLink(*someURLStr, urlReplaces)
		if err != nil {
			return nil, err
		}
		stURL := storeURL{
			ID:  *ID,
			URL: someURL.String(),
		}
		return &stURL, nil
	}
	return nil, errors.Wrap(ErrNoURL, "Store URL")
}

type storeURLs struct {
	Ios     *storeURL `json:"ios,omitempty"`
	Android *storeURL `json:"android,omitempty"`
}

func newStoreURLs(iosAppID, iosAppURL, andrAppID, andrAppURL *string, urlReplaces map[string]string) (*storeURLs, error) {
	stURLs := storeURLs{}
	var errAndr, errIos error
	stURLs.Ios, errIos = newStoreURL(iosAppID, iosAppURL, urlReplaces)
	stURLs.Android, errAndr = newStoreURL(andrAppID, andrAppURL, urlReplaces)
	if stURLs.Ios != nil || stURLs.Android != nil {
		return &stURLs, errorswrapper.WrapErrorSlice([]error{errAndr, errIos})
	}
	return nil, errorswrapper.WrapErrorSlice([]error{errAndr, errIos})
}

type phone struct {
	Value *string `json:"value,omitempty"`
	Text  *string `json:"text,omitempty"`
}

func newPhone(value *string, text *string) *phone {
	if value != nil && *value != "" {
		phone := phone{
			Value: value,
			Text:  text,
		}
		return &phone
	}
	return nil
}
