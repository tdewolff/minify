package minify

import (
	"bytes"
	"io"
)

type Mapping struct {
	src             int32
	srcLine, srcCol int32
	dstLine, dstCol int32
}

type SourceMap struct {
	sources  [][]byte
	names    [][]byte
	mappings []Mapping
}

func NewSourceMap(sources [][]byte) *SourceMap {
	return &SourceMap{
		sources: sources,
	}
}

func (sm *SourceMap) Add(src, srcLine, srcCol, dstLine, dstCol int, name []byte) {
	sm.mappings = append(sm.mappings, Mapping{
		src:     int32(src),
		srcLine: int32(srcLine),
		srcCol:  int32(srcCol),
		dstLine: int32(dstLine),
		dstCol:  int32(dstCol),
	})
	sm.names = append(sm.names, name)
}

func (sm *SourceMap) Write(w io.Writer) error {
	w.Write([]byte(`{"version":3,"sources":["`))
	for i, source := range sm.sources {
		if i != 0 {
			w.Write([]byte(`","`))
		}
		w.Write(source)
	}
	w.Write([]byte(`"],"names":["`))
	for i, name := range sm.names {
		if i != 0 {
			w.Write([]byte(`","`))
		}
		w.Write(name)
	}
	w.Write([]byte(`"],"mappings":"`))
	var prevSrc, prevSrcLine, prevSrcCol, prevDstLine int32
	for i, mapping := range sm.mappings {
		if prevDstLine != mapping.dstLine {
			w.Write(bytes.Repeat([]byte(`;`), int(mapping.dstLine-prevDstLine)))
			prevDstLine = mapping.dstLine
		} else if i != 0 {
			w.Write([]byte(`,`))
		}
		sm.WriteVLQ(mapping.dstCol)
		if prevSrc != mapping.src {
			sm.WriteVLQ(mapping.src - prevSrc)
			prevSrc = mapping.src
		}
		if prevSrcLine != mapping.srcLine {
			sm.WriteVLQ(mapping.srcLine - prevSrcLine)
			prevSrcLine = mapping.srcLine
		}
		if prevSrcCol != mapping.srcCol {
			sm.WriteVLQ(mapping.srcCol - prevSrcCol)
			prevSrcCol = mapping.srcCol
		}
	}
	_, err := w.Write([]byte(`"}`))
	return err
}

func (sm *SourceMap) WriteVLQ(val int32) {
	// TODO
}
