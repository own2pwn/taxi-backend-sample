package product

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddEmptyTariff(t *testing.T) {
	p := Product{}
	p.AddTariff("")
	assert.Nil(t, p.Tariffs)
}

func TestAddTariff(t *testing.T) {
	p := Product{}
	p.AddTariff("test")
	assert.Equal(t, []string{"test"}, p.Tariffs)
}
