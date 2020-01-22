package cli

import "fmt"

type Context struct {
	app     *App
	command *Command

	// parent is the context scope of the parent command
	parent *Context

	positionalArgs []string
	parsedFlags    map[string]Flag
	scopeFlags     map[string]Flag
	scopeCommands  map[string]*Command
}

func NewContext(app *App, parent *Context, cmd *Command) (*Context, error) {
	var flags []Flag
	ctx := &Context{
		app:     app,
		command: cmd,
		parent:  parent,

		parsedFlags:   make(map[string]Flag),
		scopeFlags:    make(map[string]Flag),
		scopeCommands: make(map[string]*Command),
	}

	if app == nil {
		return nil, fmt.Errorf(
			"NewContext invalid argument: missing app")
	}

	if cmd == nil {
		// Root scope
		flags = ctx.app.Flags
		for _, cmd := range ctx.app.Commands {
			if err := cmd.Validate(); err != nil {
				return nil, err
			}
			ctx.scopeCommands[cmd.Name] = cmd
		}
	} else {
		// Command scope
		flags = cmd.Flags
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

	for _, flag := range flags {
		if err := flag.Validate(); err != nil {
			return nil, err
		}
		if flag == nil {
			return nil, fmt.Errorf("NewContext nil flag detected!")
		}
		props := flag.GetProperties()
		ctx.scopeFlags[props.Name] = flag
		if props.Required {
			app.requiredFlags[props.Name] = flag
		}
		if props.Char != rune(0) {
			ctx.scopeFlags[string(props.Char)] = flag
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
			if value, ok := flag.GetValue().(string); ok {
				ret = value
			} else {
				break
			}
			for k, v := range c.parsedFlags {
				fmt.Printf("%s: %s\n", k, v)
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
			if value, ok := flag.GetValue().(int); ok {
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
			if value, ok := flag.GetValue().(bool); ok {
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
