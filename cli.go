package cli

import (
	"fmt"
	"strings"
)

type App struct {
	Name        string
	Author      string
	Version     [3]uint
	Usage       string
	Description string

	Action   func(ctx *Context) error
	Flags    []Flag
	Commands []*Command

	requiredFlags map[string]Flag
}

func (a *App) PrintHelp() {
	return // TODO
}

func (a *App) Run(args []string) error {
	a.requiredFlags = make(map[string]Flag)
	ctx, err := a.parseArgs(args)
	if err != nil {
		fmt.Println("Error: " + err.Error())
		if ctx == nil {
			a.PrintHelp()
		} else if ctx.command == nil {
			ctx.app.PrintHelp()
		} else {
			a.PrintHelp()
		}
		return err
	}

	if len(ctx.app.requiredFlags) > 0 {
		missingFlags := "[ "
		for k, _ := range ctx.app.requiredFlags {
			missingFlags += k + " "
			delete(ctx.app.requiredFlags, k)
		}
		missingFlags += "]"
		return fmt.Errorf(
			"The following flags are required but missing: %s",
			missingFlags)
	}

	if ctx.command == nil {
		if ctx.app.Action == nil {
			ctx.app.PrintHelp()
			return nil
		} else {
			return ctx.app.Action(ctx)
		}
	} else if ctx.command.Action == nil {
		ctx.command.PrintHelp()
		return nil
	}

	return ctx.command.Action(ctx)
}

// parseArgs parses all passed arguments and on success returns the context
// of the inner command scope.
func (app *App) parseArgs(args []string) (*Context, error) {
	var lastFlag Flag
	ctx, err := NewContext(app, nil, nil)
	if err != nil {
		return nil, err
	}

	for i, arg := range args {
		// Flag from last iteration - try to assign arg as value.
		if lastFlag != nil {
			if strings.HasPrefix(arg, "-") {
				break
			}
			switch lastFlag.(type) {
			case *BoolFlag:
				// Try to set argument, if error set to true
				if err := lastFlag.Set(arg); err != nil {
					lastFlag.Set("true")
				}

			default:
				// StringFlag & IntFlag:
				if err := lastFlag.Set(arg); err != nil {
					return nil, err
				}
			}
			ctx.parsedFlags[lastFlag.GetName()] = lastFlag
			lastFlag = nil
			continue
		}

		ret, err := parseArg(arg, ctx)
		if err != nil {
			return nil, err
		}
		switch ret.(type) {
		case Flag:
			if bf, ok := ret.(*BoolFlag); ok {
				bf.Value = true
			}
			lastFlag = ret.(Flag)

		case *Command:
			cmd := ret.(*Command)
			ctx, err = NewContext(app, ctx, cmd)
			if err != nil {
				return nil, err
			}

		case string:
			p := ret.(string)
			if p == "--" {
				ctx.positionalArgs = append(
					ctx.positionalArgs, args[i:]...)
				return ctx, nil
			}
			ctx.positionalArgs = append(ctx.positionalArgs, p)
		}
	}

	switch lastFlag.(type) {
	case *StringFlag, *IntFlag:
		return nil, fmt.Errorf(
			"The following flag is missing a value: %s",
			lastFlag.GetName())
	}

	return ctx, nil
}

func parseArg(arg string, ctx *Context) (interface{}, error) {
	var ret interface{}

	if strings.HasPrefix(arg, "--") {
		if arg == "--" {
			return arg, nil
		}
		flagName := strings.TrimPrefix(arg, "--")
		flagKeyVal := strings.SplitN(flagName, "=", 2)
		flagAddr, ok := ctx.scopeFlags[flagKeyVal[0]]
		if !ok {
			return nil, fmt.Errorf("unrecognized flag: %s", arg)
		}
		switch len(flagKeyVal) {
		// Flag has the form --flag=value
		case 2:
			if err := flagAddr.Set(flagKeyVal[1]); err != nil {
				return nil, err
			}
			ctx.parsedFlags[flagKeyVal[0]] = flagAddr
			ret = nil

		// Flag has the form --flag [value]
		case 1:
			ret = flagAddr
		}
		delete(ctx.app.requiredFlags,
			flagAddr.GetName())
		return ret, nil

	} else if strings.HasPrefix(arg, "-") {
		// Handle short flag (possibly compound)
		if arg == "-" {
			// Treat single hyphen as positional argument
			return arg, nil
		}
		charFlags := strings.TrimPrefix(arg, "-")
		rawFlags := strings.Split(charFlags, "")
		for _, char := range rawFlags[:len(rawFlags)-1] {
			flag, ok := ctx.scopeFlags[char]
			if !ok {
				return nil, fmt.Errorf(
					"unrecognized option: %s", char)
			}
			if bf, ok := flag.(*BoolFlag); ok {
				bf.Value = true
			} else {
				return nil, fmt.Errorf(
					"non-boolean flag '-%s' cannot "+
						"be used in a compound "+
						"expression '%s'",
					char, arg)
			}
			flagName := flag.GetName()
			delete(ctx.app.requiredFlags, flagName)
			if _, ok := ctx.parsedFlags[flagName]; ok {
				return nil, fmt.Errorf(
					"flag provided more than once: " +
						flagName)
			}
			ctx.parsedFlags[flagName] = flag
		}
		// Last flag of a compound expression can be whatever
		char := rawFlags[len(rawFlags)-1]
		if flag, ok := ctx.scopeFlags[char]; ok {
			flagName := flag.GetName()
			if _, ok := ctx.parsedFlags[flagName]; ok {
				return nil, fmt.Errorf(
					"flag provided more than once: " +
						flagName)
			}
			delete(ctx.app.requiredFlags, flagName)
			if bf, ok := flag.(*BoolFlag); ok {
				bf.Value = true
				ctx.parsedFlags[bf.Name] = flag
				return flag, nil
			}
			return flag, nil
		}
		return nil, fmt.Errorf("unrecognized option: %s",
			rawFlags[len(rawFlags)-1])
	} else if cmd, ok := ctx.scopeCommands[arg]; ok {
		// Check if arg is a command
		return cmd, nil
	}
	return arg, nil
}
