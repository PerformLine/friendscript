package scripting

import (
	"fmt"
	"strings"

	"github.com/PerformLine/go-stockutil/stringutil"
)

type Command struct {
	statement             *Statement
	node                  *node32
	ctx                   *Context
	overrideResultVarName string
}

func (self *Command) SourceContext() *Context {
	if self.ctx == nil {
		mod, name := self.Name()

		self.ctx = &Context{
			Type:                CommandContext,
			Label:               fmt.Sprintf("%s::%s", mod, name),
			Script:              self.statement.Script(),
			Filename:            self.statement.Script().Filename(),
			Parent:              self.statement.SourceContext(),
			AbsoluteStartOffset: int(self.node.begin),
			Length:              int(self.node.end - self.node.begin),
		}
	}

	return self.ctx
}

func (self *Command) SetOutputNameOverride(name string) {
	self.overrideResultVarName = name
}

func (self *Command) String() string {
	first, second, err := self.Args()
	var f, s string

	if err == nil {
		if first != nil {
			f = fmt.Sprintf("%v", first)
		}

		if second != nil {
			s = `{ ... }`
		}
	} else {
		s = fmt.Sprintf("!(ARGERR %v)", err)
	}

	module, name := self.Name()

	if module != UnqualifiedModuleName {
		name = module + `::` + name
	}

	if output := self.OutputName(); output == `` {
		return strings.Join(
			strings.Fields(
				fmt.Sprintf(
					fmt.Sprintf("Command %v %v %v",
						name,
						f,
						s,
					),
				),
			),
			` `,
		)
	} else {
		return strings.Join(
			strings.Fields(
				fmt.Sprintf(
					fmt.Sprintf("Command %v %v %v -> $%v",
						name,
						f,
						s,
						output,
					),
				),
			),
			` `,
		)
	}
}

func (self *Command) Script() *Friendscript {
	return self.statement.Script()
}

// Return the name of the module the command resides in and the command name.
func (self *Command) Name() (string, string) {
	ident := self.node.first(ruleCommandName)
	cmdname := self.statement.raw(ident)
	modname := UnqualifiedModuleName

	if strings.Contains(cmdname, `::`) {
		modname, cmdname = stringutil.SplitPair(cmdname, `::`)
	}

	return modname, cmdname
}

// Return the first and (optional) second arguments to a command.  If the first argument is nil, but the second
// argument is not, then the second argument will be returned as first, and the second argument will return as
// nil.  In this way, nil first arguments are collapsed and omitted.
func (self *Command) Args() (first interface{}, second map[string]interface{}, argerr error) {
	if firstNode := self.node.first(ruleCommandFirstArg); firstNode != nil {
		if variable := firstNode.firstChild(); variable != nil && variable.rule() == ruleVariable {
			if v, err := self.statement.resolveVariable(variable); err == nil {
				first = v
			} else {
				argerr = err
			}
		} else if v, err := self.statement.parseValue(firstNode); err == nil {
			first = v
		} else {
			argerr = err
		}
	}

	if secondArg := self.node.first(ruleCommandSecondArg); secondArg != nil {
		if s, err := self.statement.parseObject(secondArg.first(ruleObject)); err == nil {
			if len(s) > 0 {
				second = s
			}
		} else {
			argerr = err
		}
	}

	if first == nil && second != nil {
		first = second
		second = nil
	}

	return
}

// Return the name of the variable that command output should be stored in.
func (self *Command) OutputName() string {
	if self.overrideResultVarName != `` {
		return self.overrideResultVarName
	}

	result := self.node.first(ruleCommandResultAssignment)

	if result != nil {
		if varname := result.first(ruleVariableNameSequence); varname != nil {
			return self.statement.raw(varname)
		}
	}

	return ``
}
