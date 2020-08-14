package js

import (
	"bytes"
	"sort"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
)

type renamer struct {
	ast      *js.AST
	reserved map[string]struct{}
	rename   bool
}

func newRenamer(ast *js.AST, undeclared js.VarArray, rename bool) *renamer {
	reserved := make(map[string]struct{}, len(js.Keywords)+len(js.Globals))
	for name, _ := range js.Keywords {
		reserved[name] = struct{}{}
	}
	for name, _ := range js.Globals {
		reserved[name] = struct{}{}
	}
	return &renamer{
		ast:      ast,
		reserved: reserved,
		rename:   rename,
	}
}

func (r *renamer) renameScope(scope js.Scope) {
	if !r.rename {
		return
	}

	rename := []byte("`") // so that the next is 'a'
	sort.Sort(js.VarsByUses(scope.Declared))
	for _, v := range scope.Declared {
		if v.Link == nil {
			rename = r.next(rename)
			for r.isReserved(rename, scope.Undeclared) {
				rename = r.next(rename)
			}
			v.Name = parse.Copy(rename)
		}
	}
}

func (r *renamer) isReserved(name []byte, undeclared js.VarArray) bool {
	if 1 < len(name) { // there are no keywords or known globals that are one character long
		if _, ok := r.reserved[string(name)]; ok {
			return true
		}
	}
	for _, v := range undeclared {
		for v.Link != nil {
			v = v.Link
		}
		if bytes.Equal(name, v.Name) {
			return true
		}
	}
	return false
}

func (r *renamer) next(name []byte) []byte {
	// generate new names for variables where the last character is (a-zA-Z$_) and others are (a-zA-Z). Thus we can have 54 one-character names and 52*54=2808 two-character names. That is sufficient for virtually all input.
	if name[len(name)-1] == 'z' {
		name[len(name)-1] = 'A'
	} else if name[len(name)-1] == 'Z' {
		name[len(name)-1] = '_'
	} else if name[len(name)-1] == '_' {
		name[len(name)-1] = '$'
	} else if name[len(name)-1] == '$' {
		i := len(name) - 2
		for ; 0 <= i; i-- {
			if name[i] == 'Z' {
				continue // happens after 52*54=2808 variables
			} else if name[i] == 'z' {
				name[i] = 'A' // happens after 26*54=1404 variables
			} else {
				name[i]++
				break
			}
		}
		for j := i + 1; j < len(name); j++ {
			name[j] = 'a'
		}
		if i < 0 {
			name = append(name, 'a')
		}
	} else {
		name[len(name)-1]++
	}
	return name
}
