package cli

import (
	"fmt"
	"os"
	"strings"
)

type Context struct {
	App     *App
	Command *Command

	// parent is the context scope of the parent command
	parent *Context

	positionalArgs []string
	scopeFlags     map[string]*Flag
	parsedFlags    map[string]*Flag
	requiredFlags  map[string]*Flag
	scopeCommands  map[string]*Command
}

func NewContext(app *App, parent *Context, cmd *Command) (*Context, error) {
	var flags *[]*Flag
	ctx := &Context{
		App:     app,
		Command: cmd,
		parent:  parent,

		parsedFlags:   make(map[string]*Flag),
		requiredFlags: make(map[string]*Flag),
		scopeFlags:    make(map[string]*Flag),
		scopeCommands: make(map[string]*Command),
	}

	if app == nil {
		return nil, fmt.Errorf(
			"NewContext invalid argument: missing app")
	}

	if cmd == nil {
		// Root scope
		flags = &ctx.App.Flags
		if !ctx.App.DisableHelpCommand && len(ctx.App.Commands) > 0 {
			ctx.App.Commands = append(ctx.App.Commands, HelpCommand)
			ctx.scopeCommands[HelpCommand.Name] = HelpCommand
		}
		for _, cmd := range ctx.App.Commands {
			if err := cmd.Validate(); err != nil {
				return nil, err
			}
			ctx.scopeCommands[cmd.Name] = cmd
		}
	} else {
		// Command scope

		if !ctx.App.DisableHelpCommand &&
			// Add default help command
			len(ctx.Command.SubCommands) > 0 {
			ctx.Command.SubCommands = append(
				ctx.Command.SubCommands, HelpCommand)
		}

		flags = &cmd.Flags
		if cmd.InheritParentFlags {
			for k, v := range parent.scopeFlags {
				ctx.scopeFlags[k] = v
			}
		}
		for _, subCmd := range cmd.SubCommands {
			if err := cmd.Validate(); err != nil {
				return nil, err
			}
			ctx.scopeCommands[subCmd.Name] = subCmd
		}
	}
	if !ctx.App.DisableHelpOption && !(ctx.Command != nil &&
		(ctx.Command.InheritParentFlags ||
			ctx.Command.Name == "help")) {
		if flags != nil {
			*flags = append(*flags, HelpOption)
		} else {
			*flags = []*Flag{HelpOption}
		}
		ctx.scopeFlags[HelpOption.Name] = HelpOption
	}

	for _, flag := range *flags {
		flag.init()
		if err := flag.Validate(); err != nil {
			return nil, err
		}
		if flag == nil {
			return nil, fmt.Errorf("NewContext nil flag detected!")
		}
		ctx.scopeFlags[flag.Name] = flag
		if flag.Required {
			ctx.requiredFlags[flag.Name] = flag
		}
		if flag.Char != rune(0) {
			ctx.scopeFlags[string(flag.Char)] = flag
		}
	}

	return ctx, nil
}

// GetParent returns the parent context
func (ctx *Context) GetParent() *Context {
	return ctx.parent
}

// GetPositionals returns the positional arguments under the scope of the
// context.
func (ctx *Context) GetPositionals() []string {
	return ctx.positionalArgs
}

// String gets the value of the flag with the given name and returns whether the
// flag is set.
func (ctx *Context) String(name string) (string, bool) {
	var ret string = ""
	var isSet bool = false

	for c := ctx; c != nil; c = c.parent {
		if flag, ok := c.scopeFlags[name]; ok {
			if value, ok := flag.value.(string); ok {
				ret = value
			} else {
				break
			}
			if _, ok := c.parsedFlags[name]; ok {
				isSet = true
				break
			}
		}
	}
	return ret, isSet
}

// Int gets the value of the flag with the given name and returns whether the
// flag is set
func (ctx *Context) Int(name string) (int, bool) {
	var ret int = 0
	var isSet bool = false

	for c := ctx; c != nil; c = c.parent {
		if flag, ok := c.scopeFlags[name]; ok {
			if value, ok := flag.value.(int); ok {
				ret = value
			} else {
				break
			}
			if _, ok := c.parsedFlags[name]; ok {
				isSet = true
				break
			}
		}
	}
	return ret, isSet
}

// Bool gets the value of the flag with the given name and returns whether the
// flag is set.
func (ctx *Context) Bool(name string) (bool, bool) {
	var ret bool = false
	var isSet bool = false

	for c := ctx; c != nil; c = c.parent {
		if flag, ok := c.scopeFlags[name]; ok {
			if value, ok := flag.value.(bool); ok {
				ret = value
			} else {
				break
			}
			if _, ok := c.parsedFlags[name]; ok {
				isSet = true
				break
			}
		}
	}
	return ret, isSet
}

// Int gets the value of the flag with the given name and returns whether the
// flag is set
func (ctx *Context) Float(name string) (float64, bool) {
	var ret float64 = 0
	var isSet bool = false

	for c := ctx; c != nil; c = c.parent {
		if flag, ok := c.scopeFlags[name]; ok {
			if value, ok := flag.value.(float64); ok {
				ret = value
			} else {
				break
			}
			if _, ok := c.parsedFlags[name]; ok {
				isSet = true
				break
			}
		}
	}
	return ret, isSet
}

func (ctx *Context) Set(flag, value string) error {
	var err error
	if flag, ok := ctx.scopeFlags[flag]; ok {
		err = flag.Set(value)
		ctx.parsedFlags[flag.Name] = flag
	} else {
		err = fmt.Errorf("flag not defined")
	}
	return err
}

func (ctx *Context) assignFlag(arg string, flag *Flag) (bool, error) {
	// Ignore this check for bool and string flags
	// -- boolean flags default to true
	// -- string flags treat the next argument as a regardless string
	if flag.Type != Bool && flag.Type != String {
		// Check that the value is not a flag or command
		var argAsFlag string
		if len(arg) == 2 {
			argAsFlag = strings.TrimPrefix(arg, "-")
		} else {
			argAsFlag = strings.TrimPrefix(arg, "--")
		}
		_, isFlag := ctx.scopeFlags[argAsFlag]
		if isFlag {
			return false, fmt.Errorf(
				"error parsing arguments: "+
					"expected value of type %s, "+
					"found flag: %s",
				flag.Type, arg)
		}
		_, isCommand := ctx.scopeCommands[arg]
		if isCommand {
			return false, fmt.Errorf(
				"error parsing arguments: "+
					"expected value of type %s, "+
					"found command: %s",
				flag.Type, arg)
		}
	}
	if err := flag.Set(arg); err != nil {
		if flag.Type == Bool {
			err = nil
		}
		return false, err
	}
	ctx.parsedFlags[flag.Name] = flag
	return true, nil
}

// Free releases all internal lookup maps for the garbage collector after free
// is called this context can't be used.
func (ctx *Context) Free() {
	var p *Context
	for p = ctx; p != nil; p = p.parent {
		p.parsedFlags = nil
		p.positionalArgs = nil
		p.requiredFlags = nil
		p.scopeCommands = nil
		p.scopeFlags = nil
	}
}

func (ctx *Context) PrintHelp() error {
	helpPrinter := NewHelpPrinter(ctx, os.Stdout)
	return helpPrinter.PrintHelp()
}

func (ctx *Context) PrintUsage() error {
	helpPrinter := NewHelpPrinter(ctx, os.Stdout)
	return helpPrinter.PrintUsage()
}
