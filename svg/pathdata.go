package svg

import (
	"strconv"
	"unsafe"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse"
)

type pathData struct {
	x, y   float64
	coords [][]byte

	alterBuffer []byte
	coordBuffer []byte
}

func shortenPathData(b []byte, p *pathData) []byte {
	var x0, y0 float64
	var cmd byte

	p.x, p.y = 0.0, 0.0
	p.coords = p.coords[:0]

	j := 0
	for i := 0; i < len(b); i++ {
		c := b[i]
		if c == ' ' || c == ',' || c == '\n' || c == '\r' || c == '\t' {
			continue
		} else if c >= 'A' { // any command
			if cmd == 0 {
				cmd = c
			} else if c != cmd {
				x1, y1 := x0, y0
				if cmd == 'M' {
					x1 = toFloat(p.coords[len(p.coords)-2])
					y1 = toFloat(p.coords[len(p.coords)-1])
				} else if cmd == 'm' {
					x1 += toFloat(p.coords[len(p.coords)-2])
					y1 += toFloat(p.coords[len(p.coords)-1])
				}
				j += p.copyInstruction(b[j:], cmd)
				if cmd == 'M' || cmd == 'm' || cmd == 'Z' || cmd == 'z' {
					x0 = x1
					y0 = y1
					p.x = x0
					p.y = y0
				}
				cmd = c
				p.coords = p.coords[:0]
			}
		} else if n := parse.Number(b[i:]); n > 0 {
			p.coords = append(p.coords, minify.Number(b[i:i+n]))
			i += n - 1
		}
	}
	j += p.copyInstruction(b[j:], cmd)
	return b[:j]
}

func (p *pathData) copyInstruction(b []byte, cmd byte) int {
	n := len(p.coords)
	cmdIsRelative := cmd >= 'a'

	// get new cursor coordinates
	ax, ay := p.x, p.y
	if n >= 2 && (cmd == 'L' || cmd == 'l' || cmd == 'C' || cmd == 'c' || cmd == 'S' || cmd == 's' || cmd == 'Q' || cmd == 'q' || cmd == 'T' || cmd == 't' || cmd == 'A' || cmd == 'a') {
		ax = toFloat(p.coords[n-2])
		ay = toFloat(p.coords[n-1])
	} else if n >= 1 && (cmd == 'H' || cmd == 'h' || cmd == 'V' || cmd == 'v') {
		if cmd == 'H' || cmd == 'h' {
			ax = toFloat(p.coords[n-1])
		} else {
			ay = toFloat(p.coords[n-1])
		}
	}

	// make an alternative path with absolute/relative altered
	cmdAlter := cmd - 'A' + 'a'
	dx, dy := -p.x, -p.y
	if cmdIsRelative {
		cmdAlter = cmd - 'a' + 'A'
		dx, dy = p.x, p.y
	}
	p.alterBuffer = p.copyAlteredInstruction(p.alterBuffer[:0], cmdAlter, dx, dy)

	// choose shortest, relative or absolute path?
	j := p.copyCurrentInstruction(b, cmd)
	jAlter := len(p.alterBuffer)
	if jAlter < j {
		j = jAlter
		copy(b, p.alterBuffer)
	}

	// set new cursor coordinates
	if cmdIsRelative {
		p.x += ax
		p.y += ay
	} else {
		p.x = ax
		p.y = ay
	}
	return j
}

func (p *pathData) copyCurrentInstruction(b []byte, cmd byte) int {
	prevDigit := false
	prevDigitRequiresSpace := true

	b[0] = cmd
	j := 1
	for _, coord := range p.coords {
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
	return j
}

func (p *pathData) copyAlteredInstruction(b []byte, cmd byte, dx, dy float64) []byte {
	prevDigit := false
	prevDigitRequiresSpace := true

	b = append(b, cmd)
	for i, coord := range p.coords {
		f := toFloat(coord)
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
		p.coordBuffer = strconv.AppendFloat(p.coordBuffer[:0], f, 'f', -1, 32)
		p.coordBuffer = minify.Number(p.coordBuffer)

		if prevDigit && (p.coordBuffer[0] >= '0' && p.coordBuffer[0] <= '9' || p.coordBuffer[0] == '.' && prevDigitRequiresSpace) {
			b = append(b, ' ')
		}
		prevDigit = true
		prevDigitRequiresSpace = true
		for _, c := range p.coordBuffer {
			if c == '.' || c == 'e' || c == 'E' {
				prevDigitRequiresSpace = false
				break
			}
		}
		b = append(b, p.coordBuffer...)
	}
	return b
}

func toFloat(b []byte) float64 {
	s := *(*string)(unsafe.Pointer(&b))
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return f
}
