package xml

// EntitiesMap are all named character entities.
var EntitiesMap = map[string][]byte{
	"apos": []byte("'"),
	"quot": []byte("\""),
}

// AttrRevEntitiesMap keeps whitespace character references; decoding them to
// literal tab/LF/CR would let attribute-value normalization collapse them to a space.
var AttrRevEntitiesMap = map[byte][]byte{
	'\t': []byte("&#9;"),
	'\n': []byte("&#10;"),
	'\r': []byte("&#13;"),
}

// TextRevEntitiesMap is a map of escapes.
var TextRevEntitiesMap = map[byte][]byte{
	'<': []byte("&lt;"),
	'>': []byte("&gt;"),
	'&': []byte("&amp;"),
}
