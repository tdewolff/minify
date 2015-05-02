package svg // import "github.com/tdewolff/minify/svg"

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashTable(t *testing.T) {
	assert.Equal(t, ToHash([]byte("svg")), Svg, "'svg' must resolve to hash.Svg")
	assert.Equal(t, "svg", Svg.String(), "hash.Svg must resolve to 'svg'")
}
