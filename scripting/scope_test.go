package scripting

import (
	"testing"

	"github.com/ghetzel/testify/require"
)

func TestInterpolate(t *testing.T) {
	assert := require.New(t)
	scope := NewScope(nil)
	scope.Set(`x`, 1)
	scope.Set(`y`, 2)
	scope.Set(`z`, 3)
	scope.Set(`a`, 15155870)

	assert.Equal(int(1), scope.Get(`x`))
	assert.Equal(int(2), scope.Get(`y`))
	assert.Equal(int(3), scope.Get(`z`))
	assert.Equal(int(15155870), scope.Get(`a`))
	assert.Equal(`test test 1 2 3 15155870`, scope.Interpolate(`test test {x} {y} {z} {a}`))
}
