package svg

import (
	strconvStdlib "strconv"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse"
	"github.com/tdewolff/strconv"
)

type PathData struct {
	x, y        float64
	coords      [][]byte
	coordFloats []float64

	altBuffer   []byte
	coordBuffer []byte
}

func ShortenPathData(b []byte, p *PathData) []byte {
	var x0, y0 float64
	var cmd byte

	p.x, p.y = 0.0, 0.0
	p.coords = p.coords[:0]
	p.coordFloats = p.coordFloats[:0]

	j := 0
	for i := 0; i < len(b); i++ {
		c := b[i]
		if c == ' ' || c == ',' || c == '\n' || c == '\r' || c == '\t' {
			continue
		} else if c >= 'A' && (cmd == 0 || cmd != c) { // any command
			if cmd != 0 {
				j += p.copyInstruction(b[j:], cmd)
				if cmd == 'M' || cmd == 'm' {
					x0 = p.x
					y0 = p.y
				} else if cmd == 'Z' || cmd == 'z' {
					p.x = x0
					p.y = y0
				}
			}
			cmd = c
			p.coords = p.coords[:0]
			p.coordFloats = p.coordFloats[:0]
		} else if n := parse.Number(b[i:]); n > 0 {
			f, _ := strconv.ParseFloat(b[i : i+n])
			p.coords = append(p.coords, b[i:i+n])
			p.coordFloats = append(p.coordFloats, f)
			i += n - 1
		}
	}
	j += p.copyInstruction(b[j:], cmd)
	return b[:j]
}

func (p *PathData) copyInstruction(b []byte, cmd byte) int {
	n := len(p.coords)
	isRelativeCmd := cmd >= 'a'

	// get new cursor coordinates
	ax, ay := p.x, p.y
	if n >= 2 && (cmd == 'M' || cmd == 'm' || cmd == 'L' || cmd == 'l' || cmd == 'C' || cmd == 'c' || cmd == 'S' || cmd == 's' || cmd == 'Q' || cmd == 'q' || cmd == 'T' || cmd == 't' || cmd == 'A' || cmd == 'a') {
		ax = p.coordFloats[n-2]
		ay = p.coordFloats[n-1]
	} else if n >= 1 && (cmd == 'H' || cmd == 'h' || cmd == 'V' || cmd == 'v') {
		if cmd == 'H' || cmd == 'h' {
			ax = p.coordFloats[n-1]
		} else {
			ay = p.coordFloats[n-1]
		}
	} else if cmd == 'Z' || cmd == 'z' {
		b[0] = 'z'
		return 1
	} else {
		return 0
	}

	// make a current and alternated path with absolute/relative altered
	b = p.shortenCurPosInstruction(b, cmd)
	if isRelativeCmd {
		p.altBuffer = p.shortenAltPosInstruction(p.altBuffer[:0], cmd-'a'+'A', p.x, p.y)
	} else {
		p.altBuffer = p.shortenAltPosInstruction(p.altBuffer[:0], cmd-'A'+'a', -p.x, -p.y)
	}

	// choose shortest, relative or absolute path?
	if len(p.altBuffer) < len(b) {
		copy(b, p.altBuffer)
		b = b[:len(p.altBuffer)]
	}

	// set new cursor coordinates
	if isRelativeCmd {
		p.x += ax
		p.y += ay
	} else {
		p.x = ax
		p.y = ay
	}
	return len(b)
}

func (p *PathData) shortenCurPosInstruction(b []byte, cmd byte) []byte {
	prevDigit := false
	prevDigitRequiresSpace := true

	b[0] = cmd
	j := 1
	for _, coord := range p.coords {
		coord := minify.Number(coord)
		if prevDigit && (coord[0] >= '0' && coord[0] <= '9' || coord[0] == '.' && prevDigitRequiresSpace) {
			b[j] = ' '
			j++
		}
		prevDigit = true
		prevDigitRequiresSpace = true
		for _, c := range coord {
			if c == '.' || c == 'e' || c == 'E' {
				prevDigitRequiresSpace = false
				break
			}
		}
		j += copy(b[j:], coord)
	}
	return b[:j]
}

func (p *PathData) shortenAltPosInstruction(b []byte, cmd byte, dx, dy float64) []byte {
	prevDigit := false
	prevDigitRequiresSpace := true

	b = append(b, cmd)
	for i, f := range p.coordFloats {
		if cmd == 'L' || cmd == 'l' || cmd == 'C' || cmd == 'c' || cmd == 'S' || cmd == 's' || cmd == 'Q' || cmd == 'q' || cmd == 'T' || cmd == 't' || cmd == 'M' || cmd == 'm' {
			if i%2 == 0 {
				f += dx
			} else {
				f += dy
			}
		} else if cmd == 'H' || cmd == 'h' {
			f += dx
		} else if cmd == 'V' || cmd == 'v' {
			f += dy
		} else if cmd == 'A' || cmd == 'a' {
			if i%7 == 5 {
				f += dx
			} else if i%7 == 6 {
				f += dy
			}
		} else {
			continue
		}

		coord, ok := strconv.AppendFloat(p.coordBuffer[:0], f, 6)
		p.coordBuffer = coord // keep memory
		if !ok {
			p.coordBuffer = strconvStdlib.AppendFloat(p.coordBuffer[:0], f, 'g', 6, 64)
			coord = minify.Number(p.coordBuffer)
		}

		if prevDigit && (coord[0] >= '0' && coord[0] <= '9' || coord[0] == '.' && prevDigitRequiresSpace) {
			b = append(b, ' ')
		}
		prevDigit = true
		prevDigitRequiresSpace = true
		for _, c := range coord {
			if c == '.' || c == 'e' || c == 'E' {
				prevDigitRequiresSpace = false
				break
			}
		}
		b = append(b, coord...)
	}
	return b
}
