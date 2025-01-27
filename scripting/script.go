package scripting

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/structs"
	"github.com/PerformLine/go-stockutil/maputil"
	"github.com/PerformLine/go-stockutil/mathutil"
	"github.com/PerformLine/go-stockutil/rxutil"
	"github.com/PerformLine/go-stockutil/sliceutil"
	"github.com/PerformLine/go-stockutil/stringutil"
	"github.com/PerformLine/go-stockutil/typeutil"
)

var rxPegContext = regexp.MustCompile(`(?P<message>.*) \(line (?P<line>\d+) symbol (?P<symbol>\d+)(?: - line (?P<eline>\d+) symbol (?P<esymbol>\d+))?`)
var errContextLinesBefore = 3
var errContextLinesAfter = 3

type mappable interface {
	ToMap() map[string]interface{}
}

//go:generate peg -inline friendscript.peg

type runtime struct {
	scope    *Scope
	filename string
}

type nodeFunc func(node *node32, depth int)

func Parse(input string) (*Friendscript, error) {
	structs.DefaultTagName = `json`

	fs := &Friendscript{
		Buffer: input,
		Pretty: true,
		runtime: runtime{
			scope: NewScope(nil),
		},
	}

	fs.Init()

	if err := fs.Parse(); err == nil {
		return fs, nil
	} else if strings.HasPrefix(strings.TrimSpace(err.Error()), `parse error near`) {
		return nil, fs.errorWithContext(err)
	} else {
		return nil, err
	}
}

func LoadFromFile(filename string) (*Friendscript, error) {
	if file, err := os.Open(filename); err == nil {
		defer file.Close()

		if data, err := ioutil.ReadAll(file); err == nil {
			if fs, err := Parse(string(data)); err == nil {
				fs.filename = filename

				return fs, nil
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Friendscript) SetScope(scope *Scope) {
	self.runtime.scope = scope
}

func (self *Friendscript) Filename() string {
	return self.runtime.filename
}

func (self *Friendscript) Scope() *Scope {
	return self.runtime.scope
}

func (self *Friendscript) lcp() string {
	return stringutil.LongestCommonPrefix(strings.Split(self.Buffer, "\n"))
}

func (self *Friendscript) errorWithContext(err error) error {
	raw := strings.TrimSpace(err.Error())

	if match := rxutil.Match(rxPegContext, raw); match != nil {
		line := int(stringutil.MustInteger(match.Group(`line`)))
		symbol := int(stringutil.MustInteger(match.Group(`symbol`)))

		lines := strings.Split(self.Buffer, "\n")
		lbound := int(mathutil.ClampLower(float64(line-errContextLinesBefore), 0))
		ubound := int(mathutil.ClampUpper(float64(line+errContextLinesAfter), float64(len(lines))))
		message := fmt.Sprintf("Syntax error on line %d: %v\n", line, match.Group(`message`))
		message += "\n"
		lcp := stringutil.LongestCommonPrefix(lines)

		for i := lbound; i < ubound; i++ {
			message += fmt.Sprintf("%- 4d | %v\n", i, strings.TrimPrefix(lines[i], lcp))

			if i == (line - 1) {
				sl := (symbol - 1)

				if sl < 0 {
					sl = 0
				}

				message += fmt.Sprintf("     | %s^\n", strings.Repeat(`-`, sl))
			}
		}

		return errors.New(message)
	} else {
		return err
	}
}

// Return all top-level blocks in the current script.
func (self *Friendscript) Blocks() []*Block {
	blocks := make([]*Block, 0)
	root := self.AST()

	root.traverse(func(node *node32, depth int) {
		// fmt.Printf("[% 2d] %v%v %q\n", depth, strings.Repeat(`  `, depth), node, self.s(node))

		switch node.rule() {
		case ruleStatementBlock:
			blocks = append(blocks, &Block{
				friendscript: self,
				node:         node,
			})
		}
	}, 1)

	return blocks
}

func (self *Friendscript) s(node *node32) string {
	if node != nil {
		begin := int(node.token32.begin)
		end := int(node.token32.end)
		return self.Buffer[begin:end]
	} else {
		return ``
	}
}

func (self *node32) rule() pegRule {
	return self.token32.pegRule
}

func (self *node32) first(anyOf ...pegRule) *node32 {
	return self.firstN(-1, anyOf...)
}

func (self *node32) firstN(maxdepth int, anyOf ...pegRule) *node32 {
	results := self.findN(maxdepth, anyOf...)

	if len(results) > 0 {
		return results[0]
	} else {
		return nil
	}
}

func (self *node32) firstChild(anyOf ...pegRule) *node32 {
	return self.firstN(0, anyOf...)
}

func (self *node32) find(anyOf ...pegRule) []*node32 {
	return self.findN(-1, anyOf...)
}

func (self *node32) children(anyOf ...pegRule) []*node32 {
	return self.findN(0, anyOf...)
}

func (self *node32) findN(maxdepth int, anyOf ...pegRule) []*node32 {
	results := make([]*node32, 0)

	self.traverse(func(node *node32, _ int) {
		if node == self {
			return
		}

		switch node.rule() {
		case rule_, rule__:
			return
		}

		if len(anyOf) == 0 || sliceutil.Contains(anyOf, node.rule()) {
			results = append(results, node)
		}
	}, maxdepth)

	return results
}

func (self *node32) findUntil(maxdepth int, stopRule pegRule, anyOf ...pegRule) []*node32 {
	results := make([]*node32, 0)
	hitStopRule := false

	self.traverse(func(node *node32, _ int) {
		if !hitStopRule {
			if node == self {
				return
			}

			switch node.rule() {
			case rule_, rule__:
				return
			}

			if node.rule() == stopRule {
				hitStopRule = true
				return
			}

			if len(anyOf) == 0 || sliceutil.Contains(anyOf, node.rule()) {
				results = append(results, node)
			}
		}
	}, maxdepth)

	return results
}

func (self *node32) findAfter(maxdepth int, startRule pegRule, anyOf ...pegRule) []*node32 {
	results := make([]*node32, 0)
	hitStartRule := false

	self.traverse(func(node *node32, _ int) {
		if node == self {
			return
		}

		switch node.rule() {
		case rule_, rule__:
			return
		}

		if node.rule() == startRule {
			hitStartRule = true
			return
		}

		if hitStartRule {
			if len(anyOf) == 0 || sliceutil.Contains(anyOf, node.rule()) {
				results = append(results, node)
			}
		}
	}, maxdepth)

	return results
}

func (self *node32) traverse(nodeFn nodeFunc, maxdepth int) {
	self.traverseNode(self, nodeFn, 0, maxdepth)
}

func (self *node32) traverseNode(start *node32, nodeFn nodeFunc, depth int, maxdepth int) {
	node := start

	for node != nil {
		nodeFn(node, depth)

		if node.up != nil {
			if maxdepth < 0 || depth <= maxdepth {
				self.traverseNode(node.up, nodeFn, depth+1, maxdepth)
			}
		}

		node = node.next
	}
}

func debugNode(friendscript *Friendscript, node *node32) {
	if node != nil {
		node.traverse(func(node *node32, depth int) {
			fmt.Printf("[% 2d] %v%v %q\n", depth, strings.Repeat(`  `, depth), node, friendscript.s(node))
		}, -1)
	} else {
		fmt.Printf("empty node\n")
	}
}

// return int or int64 when appropriate
func intIfYouCan(in interface{}) interface{} {
	if oF, ok := in.(float64); ok {
		if oF == float64(int(oF)) {
			return int(oF)

		} else if oF == float64(int64(oF)) {
			return int64(oF)
		}
	}

	return in
}

// if the input is a struct, convert it into a map
func mapifyStruct(in interface{}) interface{} {
	maputil.UnmarshalStructTag = `json`

	if m, ok := in.(mappable); ok {
		return m.ToMap()

	} else if b, ok := in.([]byte); ok {
		return b
	} else if typeutil.IsArray(in) {
		var elems = make([]interface{}, sliceutil.Len(in))

		// work a little to get structs/mappable objects turned into maps
		sliceutil.Each(in, func(i int, elem interface{}) error {
			elems[i] = mapifyStruct(elem)
			return nil
		})

		return elems
	} else if typeutil.IsStruct(in) {
		return maputil.M(in).MapNative()
	} else {
		return in
	}
}
