package product

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTariffEqName(t *testing.T) {
	r := record{
		ID:       1,
		RegionID: 1,
		Name:     "uber:uber:1",
	}
	assert.Equal(t, "uber", r.name())
	assert.Equal(t, "", r.tariff())
}

func TestTariffDiffName(t *testing.T) {
	r := record{
		ID:       1,
		RegionID: 1,
		Name:     "uber:uberx:1",
	}
	assert.Equal(t, "uber", r.name())
	assert.Equal(t, "uberx", r.tariff())
}
