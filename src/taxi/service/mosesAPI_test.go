package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

type mockRegionsInfo struct{}

func (m *mockRegionsInfo) GetRegionNameByID(regID int) (string, error) {
	return "nsk", nil
}

func TestMoses(t *testing.T) {
	defer gock.Off()
	testTaxiRequest := Request{
		RegionID: 14,
		Point1: Point{
			Lon:     30.723449,
			LonStr:  "30.723449",
			Lat:     46.441982,
			LatStr:  "46.441982",
			Address: "Одесса. Тестовая точка 1",
		},
		Point2: Point{
			Lon:     30.759315,
			LonStr:  "30.759315",
			Lat:     46.453352,
			LatStr:  "46.453352",
			Address: "Одесса. Тестовая точка 2",
		},
		OnlyAPI: true,
	}
	mosesService, _ := NewMosesService(new(mockRegionsInfo))
	gock.New(mosesService.mosesURL.Host).
		Post(mosesService.mosesURL.Path).
		Reply(200).
		File("_test_jsons/moses.json")
	gock.InterceptClient(testHttpClient)
	distance, times, err := mosesService.DistanceTime(testContext, testHTTPRequester, testTaxiRequest)
	assert.Nil(t, err)
	assert.Equal(t, 3474, distance)
	assert.Equal(t, int(720/60), times)
	assert.Equal(t, gock.IsDone(), true)
}

func TestMoses0(t *testing.T) {
	defer gock.Off()
	testTaxiRequest := Request{
		RegionID: 14,
		Point1: Point{
			Lon:     30.723449,
			LonStr:  "30.723449",
			Lat:     46.441982,
			LatStr:  "46.441982",
			Address: "Одесса. Тестовая точка 1",
		},
		Point2: Point{
			Lon:     30.759315,
			LonStr:  "30.759315",
			Lat:     46.453352,
			LatStr:  "46.453352",
			Address: "Одесса. Тестовая точка 2",
		},
		OnlyAPI: true,
	}
	mosesService, _ := NewMosesService(new(mockRegionsInfo))
	gock.New(mosesService.mosesURL.Host).
		Post(mosesService.mosesURL.Path).
		Reply(200).
		File("_test_jsons/moses.Empty.json")
	gock.InterceptClient(testHttpClient)
	_, _, err := mosesService.DistanceTime(testContext, testHTTPRequester, testTaxiRequest)
	assert.NotNil(t, err)
	assert.Equal(t, gock.IsDone(), true)
}

func TestMosesErr(t *testing.T) {
	defer gock.Off()
	testTaxiRequest := Request{
		RegionID: 14,
		Point1: Point{
			Lon:     30.723449,
			LonStr:  "30.723449",
			Lat:     46.441982,
			LatStr:  "46.441982",
			Address: "Одесса. Тестовая точка 1",
		},
		Point2: Point{
			Lon:     30.759315,
			LonStr:  "30.759315",
			Lat:     46.453352,
			LatStr:  "46.453352",
			Address: "Одесса. Тестовая точка 2",
		},
		OnlyAPI: true,
	}
	mosesService, _ := NewMosesService(new(mockRegionsInfo))
	gock.New(mosesService.mosesURL.Host).
		Post(mosesService.mosesURL.Path).
		Reply(200).
		BodyString("err")
	gock.InterceptClient(testHttpClient)
	_, _, err := mosesService.DistanceTime(testContext, testHTTPRequester, testTaxiRequest)
	assert.NotNil(t, err)
	assert.Equal(t, gock.IsDone(), true)
}
