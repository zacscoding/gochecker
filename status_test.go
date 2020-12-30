package gochecker

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestComponentStatus(t *testing.T) {
	s := NewComponentStatus()
	assert.Equal(t, s.status, unknown)

	s.WithUp()
	assert.Equal(t, s.status, up)
	assert.True(t, s.IsUp())

	s.WithDown()
	assert.Equal(t, s.status, down)
	assert.True(t, s.IsDown())

	s.WithDetail("key1", "value1")
	assert.Equal(t, s.details["key1"], "value1")
}
