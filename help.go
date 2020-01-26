package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

const (
	defaultWidth int = 80

	columnFraction = 0.3
	maxColumnWidth = 35

	bufferSize = 1024
)

type HelpPrinter struct {
	buf         *bytes.Buffer
	ctx         *Context
	out         io.Writer
	width       int
	columnWidth int

	// Internal writer parameters
	RightMargin int
	cursor      int
	LeftMargin  int
}

// NewHelpPrinter creates a help printer initialized with the context ctx.
// Using PrintHelp will create a help prompt based on ctx that will be written
// to out.
func NewHelpPrinter(ctx *Context, out io.Writer) *HelpPrinter {
	var width int
	if f, ok := out.(*os.File); ok {
		ws, err := getTerminalSize(int(f.Fd()))
		if err != nil {
			width = defaultWidth
		} else {
			width = int(ws[0])
		}
	}
	if width < 10 {
		width = defaultWidth
	}
	columnWidth := int(columnFraction * float64(width))
	if columnWidth > maxColumnWidth {
		columnWidth = maxColumnWidth
	}

	return &HelpPrinter{
		ctx:         ctx,
		buf:         &bytes.Buffer{},
		out:         out,
		width:       width,
		columnWidth: columnWidth,

		LeftMargin:  0,
		RightMargin: width,
	}
}

// Write function which makes the HelpPrinter conform with the io.Writer
// interface. The printer attempts to insert newlines at word boundaries and
// satisfy the margin constrains in the HelpPrinter structure.
// NOTE: The returned length is that of the bytes written to the buffer -
//       that includes indentation and inserted newlines.
func (hp *HelpPrinter) Write(p []byte) (int, error) {
	var err error
	var n int
	var N int
	var NumExtraChars int
	var pp []byte
	if hp.RightMargin <= hp.LeftMargin {
		hp.LeftMargin = 0
		hp.RightMargin = defaultWidth
	}
	for N < len(p) {
		pp = p[N:]
		if hp.cursor < hp.LeftMargin {
			n, err = fmt.Fprintf(hp.buf, "%*s",
				hp.LeftMargin-hp.cursor, "")
			hp.cursor += n
			NumExtraChars += n
			if err != nil {
				break
			}
			// Trim white-space characters
			for N < len(p) && p[N] == byte(' ') {
				N++
			}
			continue
		}
		lineSpace := hp.RightMargin - hp.cursor
		if lineSpace > len(pp) {
			lineSpace = len(pp)
		} else if lineSpace <= 0 {
			n, err := fmt.Fprintln(hp.buf)
			if err != nil {
				break
			}
			NumExtraChars += n
			hp.cursor = 0
			continue
		}
		if idx := bytes.Index(pp[:lineSpace], []byte(NewLine)); idx >= 0 {
			idx += len(NewLine)
			n, err = hp.buf.Write(pp[:idx])
			hp.cursor = 0
		} else {
			// Need to split last word
			idx = bytes.LastIndex(pp[:lineSpace], []byte(" "))
			if idx < 0 {
				idx = bytes.Index(pp, []byte(" "))
				if idx < 0 {
					idx = len(pp)
				}
				if lineSpace >= idx {
					n, err = hp.buf.Write(pp)
				} else if idx > hp.RightMargin-hp.LeftMargin {
					// Last resort, next word doesn't fit so
					// flush the remainder of the line.
					n, err = hp.buf.Write(pp[:lineSpace])
				} else {
					// Insert newline, reset cursor
					n, err = fmt.Fprintln(hp.buf)
					NumExtraChars += n
					hp.cursor = 0
					if err != nil {
						break
					}
					continue
				}
			} else {
				idx += 1
				n, err = hp.buf.Write(pp[:idx])
			}
			hp.cursor += n
		}
		N += n
		if err != nil {
			break
		}
	} // for N < len(p)
	return N + NumExtraChars, err
}

func (hp *HelpPrinter) initPrint() ([]*Flag, []*Flag, string) {
	var flags []*Flag
	var execStr string

	if hp.ctx.command == nil {
		flags = hp.ctx.app.Flags
		execStr = os.Args[0]
	} else {
		for p := hp.ctx; p != nil; p = p.parent {
			if p.command == nil {
				flags = append(flags, p.app.Flags...)
			} else {
				execStr = p.command.Name + " " + execStr
				flags = append(flags, p.command.Flags...)
				if !p.command.InheritParentFlags {
					break
				}
			}
		}
		execStr = os.Args[0] + " " + execStr
	}

	optFlags, reqFlags := getOptionalAndRequired(flags)
	return optFlags, reqFlags, execStr
}

func (hp *HelpPrinter) PrintUsage() error {
	optFlags, reqFlags, execStr := hp.initPrint()
	err := hp.writeUsage(execStr, reqFlags, optFlags)
	if err != nil {
		return err
	}
	_, err = hp.buf.WriteTo(hp.out)
	return err
}

func (hp *HelpPrinter) PrintHelp() error {
	optFlags, reqFlags, execStr := hp.initPrint()
	err := hp.writeUsage(execStr, reqFlags, optFlags)
	if err != nil {
		return err
	}
	if hp.ctx.command != nil {
		if hp.ctx.command.Description != "" {
			hp.LeftMargin = 0
			fmt.Fprintln(hp, "Description:")
			hp.LeftMargin = 2
			fmt.Fprintln(hp, NewLine+hp.ctx.command.Description+NewLine)
		}
		if len(hp.ctx.command.SubCommands) > 0 {
			err = hp.writeCommandSection(hp.ctx.command.SubCommands)
		}
	} else {
		if hp.ctx.app.Description != "" {
			hp.LeftMargin = 0
			fmt.Fprintln(hp, NewLine+"Description:")
			hp.LeftMargin = 2
			fmt.Fprintln(hp, hp.ctx.app.Description)
		}
		if len(hp.ctx.app.Commands) > 0 {
			err = hp.writeCommandSection(hp.ctx.app.Commands)
		}
	}
	if err != nil {
		return err
	}

	err = hp.writeFlagSection("Required flags", reqFlags)
	if err != nil {
		return err
	}

	err = hp.writeFlagSection("Optional flags", optFlags)

	hp.buf.WriteTo(hp.out)
	return err
}

func (hp *HelpPrinter) writeCommandSection(commands []*Command) error {
	hp.LeftMargin = 0
	_, err := fmt.Fprintln(hp, NewLine+"Commands:")
	if err != nil {
		return err
	}
	for _, cmd := range commands {
		hp.LeftMargin = 2
		_, err = fmt.Fprint(hp, cmd.Name)
		if err != nil {
			return err
		}
		hp.LeftMargin = hp.columnWidth
		_, err = fmt.Fprintln(hp, cmd.Usage)
		if err != nil {
			return err
		}
	}
	return nil
}

func (hp *HelpPrinter) writeFlagSection(section string, flags []*Flag) error {
	hp.LeftMargin = 0
	_, err := fmt.Fprint(hp, NewLine+section+":"+NewLine)
	if err != nil {
		return err
	}
	for _, flag := range flags {
		char := "/-" + string(flag.Char)
		if flag.Char == rune(0) {
			char = ""
		}
		hp.LeftMargin = 2
		metaVar := flag.MetaVar
		if metaVar == "" {
			if flag.Type != Bool {
				metaVar = "value"
			}
		}

		n, err := fmt.Fprintf(hp, "--%s%s %s  ",
			flag.Name, char, metaVar)
		if err != nil {
			return err
		}
		hp.LeftMargin = hp.columnWidth
		if n > hp.LeftMargin {
			fmt.Fprintln(hp)
		}
		fmt.Fprint(hp, flag.Usage+NewLine)
	}

	return nil
}

func (hp *HelpPrinter) writeUsage(
	execStr string,
	required, optional []*Flag,
) error {

	n, err := fmt.Fprintf(hp, "Usage: %s", execStr)
	if err != nil {
		return err
	}
	if n < hp.width {
		hp.LeftMargin = n
	}

	for _, flag := range append(required, optional...) {
		word := "--" + flag.Name
		if flag.Char != rune(0) {
			word = "-" + string(flag.Char)
		}
		if flag.MetaVar == "" {
			if flag.Type != Bool {
				word += " value"
			}
		} else {
			word = fmt.Sprintf("%s %s", word, flag.MetaVar)
		}

		if flag.Required {
			word = " " + word
		} else {
			word = " [" + word + "]"
		}
		if hp.cursor+len(word) > hp.RightMargin {
			word = NewLine + word
		}
		n, err = fmt.Fprint(hp, word)
		if err != nil {
			return err
		}
	}

	// Print commands usage, use curly braces if the commands are required
	// and square brackets otherwise.
	cmdString := " ["
	suffix := "]"
	if hp.ctx.command != nil {
		if len(hp.ctx.command.SubCommands) > 0 {
			if hp.ctx.command.Action == nil {
				cmdString = " {"
				suffix = "}"
			}
			for _, cmd := range hp.ctx.command.SubCommands {
				cmdString += cmd.Name + ","
			}
			// Remove trailing comma and replace it with suffix
			cmdString = cmdString[:len(cmdString)-1] + suffix
		}
	} else if len(hp.ctx.app.Commands) > 0 {
		if hp.ctx.app.Action == nil {
			cmdString = "}"
			suffix = "}"
		}
		for _, cmd := range hp.ctx.app.Commands {
			cmdString += cmd.Name + ","
		}
		// Remove trailing comma and replace it with suffix
		cmdString = cmdString[:len(cmdString)-1] + suffix
	}
	if len(cmdString) <= 2 {
		cmdString = ""
	}
	_, err = fmt.Fprintln(hp, cmdString)

	return err
}

func getOptionalAndRequired(flags []*Flag) ([]*Flag, []*Flag) {
	var optional []*Flag
	var required []*Flag
	var numRequired int
	var i, j int

	for _, flag := range flags {
		if flag.Required {
			numRequired++
		}
	}
	required = make([]*Flag, numRequired)
	optional = make([]*Flag, len(flags)-numRequired)
	for _, flag := range flags {
		if flag.Required {
			required[i] = flag
			i++
		} else {
			optional[j] = flag
			j++
		}
	}

	return optional, required
}
