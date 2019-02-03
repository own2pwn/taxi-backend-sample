package regionsinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/nburunova/taxi-backend-sample/src/webapi"
)

func TestUpdateRegions(t *testing.T) {
	testWebAPIRegs := []webapi.RegionInfo{
		{
			ID:   1,
			Name: "novosibirsk",
		},
	}
	newReg := NewRegionsInfo()
	updErr := newReg.Load(testWebAPIRegs)
	assert.Nil(t, updErr)
	regionName, err := newReg.GetRegionNameByID(1)
	assert.Nil(t, err)
	assert.Equal(t, "novosibirsk", regionName)
	regionName, err = newReg.GetRegionNameByID(1000)
	assert.NotNil(t, err)
}
