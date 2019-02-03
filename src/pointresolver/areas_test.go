package pointresolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadAreas(t *testing.T) {

	_, err := newAreas()
	assert.Nil(t, err)
}

func TestChernomorsk(t *testing.T) {
	areas, err := newAreas()
	assert.Nil(t, err)
	name := areas.areaNameByLatLon(46.298108, 30.648476)
	assert.Equal(t, "chernomorsk", name)
}

func TestChernomorsk2(t *testing.T) {
	areas, err := newAreas()
	assert.Nil(t, err)
	name := areas.areaNameByLatLon(46.312671, 30.674987)
	assert.Equal(t, "chernomorsk", name)
}

func TestOdessa(t *testing.T) {
	areas, err := newAreas()
	assert.Nil(t, err)
	name := areas.areaNameByLatLon(46.441982, 30.723449)
	assert.Equal(t, "", name)
}
func TestOdessa2(t *testing.T) {
	areas, err := newAreas()
	assert.Nil(t, err)
	name := areas.areaNameByLatLon(46.325653, 30.668077)
	assert.Equal(t, "", name)
}

func TestDubaiSea(t *testing.T) {
	areas, err := newAreas()
	assert.Nil(t, err)
	name := areas.areaNameByLatLon(25.677408, 54.273902)
	assert.Equal(t, "", name)
}

func TestDubaiNorth(t *testing.T) {
	areas, err := newAreas()
	assert.Nil(t, err)
	name := areas.areaNameByLatLon(25.875928, 56.077523)
	assert.Equal(t, "", name)
}

func TestDubaiBetweenDubaiAndAbuDabi(t *testing.T) {
	areas, err := newAreas()
	assert.Nil(t, err)
	name := areas.areaNameByLatLon(24.884028, 54.888327)
	assert.Equal(t, "", name)
}

func TestDubaiDubai(t *testing.T) {
	areas, err := newAreas()
	assert.Nil(t, err)
	name := areas.areaNameByLatLon(25.241601, 55.442809)
	assert.Equal(t, "dubai", name)
}

func TestDubaiAbuDabi(t *testing.T) {
	areas, err := newAreas()
	assert.Nil(t, err)
	name := areas.areaNameByLatLon(24.432748, 54.442735)
	assert.Equal(t, "dubai", name)
}
