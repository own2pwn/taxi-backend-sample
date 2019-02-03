package webapi

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/collector"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/httprequester"
	"github.com/nburunova/taxi-backend-sample/src/infrastructure/log"
	"gopkg.in/h2non/gock.v1"
)

func newTestContext() context.Context {
	testContext := context.Background()
	testContext = context.WithValue(testContext, log.CtxKeyAPIName, "test")
	testContext = context.WithValue(testContext, log.CtxKeyRegionID, 1)
	return testContext
}

var (
	testContext    = newTestContext()
	testHttpClient = &http.Client{
		Transport: &http.Transport{
			// Максимальное время бездействия до закрытия соединения; сколько времи простаивающее соединение хранится в пуле
			IdleConnTimeout: 5 * time.Second,
		},
	}
	testLogger        = log.NewEmpty()
	testCollector     = collector.NewCollector(testLogger)
	testHTTPRequester = httprequester.NewRequester(testHttpClient, testLogger, testCollector)
)

func TestWeAPIPriority(t *testing.T) {
	defer gock.Off()
	lon := float64(30.723449)
	lat := float64(46.441982)
	web, _ := NewClient(testHTTPRequester)
	gock.New(webAPIURL).
		Get(geoMethod).
		Reply(200).
		File("_test_jsons/webAPI.json")
	gock.InterceptClient(testHttpClient)
	point, err := web.GetPointInfo(testContext, lat, lon)
	assert.Nil(t, err)
	assert.Equal(t, "address_name field first priority", point.Address)
	assert.Equal(t, gock.IsDone(), true)
}

func TestWeAPIAddressField(t *testing.T) {
	defer gock.Off()
	lon := float64(30.723449)
	lat := float64(46.441982)
	web, _ := NewClient(testHTTPRequester)
	gock.New(webAPIURL).
		Get(geoMethod).
		Reply(200).
		File("_test_jsons/webAPI.json")
	gock.InterceptClient(testHttpClient)
	point, err := web.GetPointInfo(testContext, lat, lon)
	assert.Nil(t, err)
	assert.Equal(t, "address_name field first priority", point.Address)
	assert.Equal(t, gock.IsDone(), true)
}

func TestWeAPINameField(t *testing.T) {
	defer gock.Off()
	lon := float64(30.723449)
	lat := float64(46.441982)
	web, _ := NewClient(testHTTPRequester)
	gock.New(webAPIURL).
		Get(geoMethod).
		Reply(200).
		File("_test_jsons/webAPI.Name.json")
	gock.InterceptClient(testHttpClient)
	point, err := web.GetPointInfo(testContext, lat, lon)
	assert.Nil(t, err)
	assert.Equal(t, "name field second priority", point.Address)
	assert.Equal(t, gock.IsDone(), true)
}

func TestWeAPIPointSpaces(t *testing.T) {
	defer gock.Off()
	lon := float64(30.723449)
	lat := float64(46.441982)
	web, _ := NewClient(testHTTPRequester)
	gock.New(webAPIURL).
		Get(geoMethod).
		Reply(200).
		File("_test_jsons/webAPI.PointFormatSpaces.json")
	gock.InterceptClient(testHttpClient)
	point, err := web.GetPointInfo(testContext, lat, lon)
	assert.Nil(t, err)
	assert.Equal(t, "address_name field first priority", point.Address)
	assert.Equal(t, gock.IsDone(), true)
}

func TestWeAPIEmptyAddress(t *testing.T) {
	defer gock.Off()
	lon := float64(30.723449)
	lat := float64(46.441982)
	web, _ := NewClient(testHTTPRequester)
	gock.New(webAPIURL).
		Get(geoMethod).
		Reply(200).
		File("_test_jsons/webAPI.Empty.json")
	gock.InterceptClient(testHttpClient)
	_, err := web.GetPointInfo(testContext, lat, lon)
	assert.NotNil(t, err)
	assert.Equal(t, gock.IsDone(), true)
}

func TestWeAPIPointNotPoint(t *testing.T) {
	defer gock.Off()
	lon := float64(30.723449)
	lat := float64(46.441982)
	web, _ := NewClient(testHTTPRequester)
	gock.New(webAPIURL).
		Get(geoMethod).
		Reply(200).
		File("_test_jsons/webAPI.PointFormatNotPoint.json")
	gock.InterceptClient(testHttpClient)
	_, err := web.GetPointInfo(testContext, lat, lon)
	assert.NotNil(t, err)
	assert.Equal(t, gock.IsDone(), true)
}

func TestWeAPIEmptyList(t *testing.T) {
	defer gock.Off()
	lon := float64(30.723449)
	lat := float64(46.441982)
	web, _ := NewClient(testHTTPRequester)
	gock.New(webAPIURL).
		Get(geoMethod).
		Reply(200).
		File("_test_jsons/webAPI.EmptyItems.json")
	gock.InterceptClient(testHttpClient)
	_, err := web.GetPointInfo(testContext, lat, lon)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), ErrEmptyResult.Error())
	assert.Equal(t, gock.IsDone(), true)
}

func TestWeAPIErr(t *testing.T) {
	defer gock.Off()
	lon := float64(30.723449)
	lat := float64(46.441982)
	web, _ := NewClient(testHTTPRequester)
	gock.New(webAPIURL).
		Get(geoMethod).
		Reply(200).
		BodyString("err")
	_, err := web.GetPointInfo(testContext, lat, lon)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), httprequester.ErrParse.Error())
	assert.Equal(t, gock.IsDone(), true)
}

func TestWeAPIErr500(t *testing.T) {
	defer gock.Off()
	lon := float64(30.723449)
	lat := float64(46.441982)
	web, _ := NewClient(testHTTPRequester)
	gock.New(webAPIURL).
		Get(geoMethod).
		Reply(500)
	gock.InterceptClient(testHttpClient)
	_, err := web.GetPointInfo(testContext, lat, lon)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), httprequester.ErrStatusNotOK.Error())
	assert.Equal(t, gock.IsDone(), true)
}

func TestLatLonParseNoBracket(t *testing.T) {
	_, _, err := pointToLatLon("POINT 37.61077 55.750524")
	assert.NotNil(t, err)
}

func TestLatLonParseCoordFormatErr(t *testing.T) {
	_, _, err := pointToLatLon("POINT (37.6107755.750524)")
	assert.NotNil(t, err)
}

func TestUpdateRegions(t *testing.T) {
	defer gock.Off()
	web, _ := NewClient(testHTTPRequester)
	gock.New(webAPIURL).
		Get(regionListMethod).
		Reply(200).
		File("_test_jsons/regionList.json")
	gock.InterceptClient(testHttpClient)
	regs, err := web.GetRegionsList(testContext)
	assert.Nil(t, err)
	assert.NotNil(t, regs)
	assert.NotEqual(t, 0, regs[0].ID)
}
