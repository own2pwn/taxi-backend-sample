package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyResponse(t *testing.T) {
	res := newResponse(nil, nil, nil)
	assert.Nil(t, res.Meta)
	assert.Nil(t, res.Result.Else)
	assert.Nil(t, res.Result.Optimal)
}
