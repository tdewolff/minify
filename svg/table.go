package svg // import "github.com/tdewolff/minify/svg"

import "github.com/tdewolff/parse/svg"

var containerTagMap = map[svg.Hash]bool{
	svg.A:             true,
	svg.Defs:          true,
	svg.G:             true,
	svg.Marker:        true,
	svg.Mask:          true,
	svg.Missing_Glyph: true,
	svg.Pattern:       true,
	svg.Svg:           true,
	svg.Switch:        true,
	svg.Symbol:        true,
}

var colorAttrMap = map[svg.Hash]bool{
	svg.Fill:           true,
	svg.Stroke:         true,
	svg.Stop_Color:     true,
	svg.Flood_Color:    true,
	svg.Lighting_Color: true,
}
