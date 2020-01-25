package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

const (
	defaultWidth int = 80

	columnFraction = 0.34

	bufferSize = 1024
)

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

func (hp *HelpPrinter) writeFlagSection(section string, flags []*Flag) error {
	hp.indent = 0
	_, err := fmt.Fprint(hp, NewLine+section+":"+NewLine)
	if err != nil {
		return err
	}
	for _, flag := range flags {
		char := "/-" + string(flag.Char)
		if flag.Char == rune(0) {
			char = ""
		}
		hp.indent = 2
		metaVar := flag.MetaVar
		if metaVar == "" {
			if flag.Type == Bool {
				metaVar = "true/false"
			} else {
				metaVar = "value"
			}
		}

		n, err := fmt.Fprintf(hp, "--%s%s %s  ",
			flag.Name, char, metaVar)
		if err != nil {
			return err
		}
		hp.indent = hp.columnWidth
		if n > hp.indent {
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
	var idx int

	n, err := fmt.Fprintf(hp, "Usage: %s ", execStr)
	if err != nil {
		return err
	}
	idx += n
	if n < hp.width {
		hp.indent = n
	}

	for _, flag := range append(required, optional...) {
		word := "--" + flag.Name
		if flag.Char != rune(0) {
			word = "-" + string(flag.Char)
		}
		metaVar := flag.MetaVar
		if metaVar == "" {
			if flag.Type == Bool {
				metaVar = "true/false"
			} else {
				metaVar = "value"
			}
		}

		word = fmt.Sprintf("%s %s", word, metaVar)
		if flag.Required {
			word += " "
		} else {
			word = "[" + word + "] "
		}
		if hp.cursor+len(word) > hp.writeWidth {
			word = NewLine + word
		}
		n, err = fmt.Fprint(hp, word)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprint(hp, NewLine)

	return err
}

type HelpPrinter struct {
	buf         *bytes.Buffer
	ctx         *Context
	out         io.Writer
	width       int
	columnWidth int

	// Internal writer parameters
	writeWidth int
	cursor     int
	indent     int
}

func NewHelpPrinter(ctx *Context, out io.Writer) *HelpPrinter {
	var width int
	if f, ok := out.(*os.File); ok {
		ws, err := getTerminalSize(int(f.Fd()))
		if err != nil {
			width = defaultWidth
		} else {
			width = int(ws[0])
		}
	} else {
		width = defaultWidth
	}
	columnWidth := int(columnFraction * float64(width))

	return &HelpPrinter{
		ctx:         ctx,
		buf:         &bytes.Buffer{},
		out:         out,
		width:       width,
		columnWidth: columnWidth,

		writeWidth: width,
		indent:     0,
	}
}

func (hp *HelpPrinter) Write(p []byte) (int, error) {
	var err error
	var n int
	var N int
	var NumIndent int
	var pp []byte
	var indented bool = false

	for N < len(p) {
		pp = p[N:]
		if hp.cursor < hp.indent {
			n, err = fmt.Fprintf(hp.buf, "%*s",
				hp.indent-hp.cursor, "")
			hp.cursor += n
			NumIndent += n
			indented = true
			if err != nil {
				break
			}
		}
		remaining := hp.writeWidth - hp.cursor
		if remaining >= len(pp) {
			remaining = len(pp)
		} else if remaining < 0 {
			fmt.Fprintln(hp.buf)
			hp.cursor = 0
			continue
		}
		idx := bytes.Index(pp[:remaining], []byte(NewLine))
		if idx >= 0 {
			idx += len(NewLine)
			n, err = hp.buf.Write(pp[:idx])
			hp.cursor = 0
		} else {
			idx = bytes.LastIndex(pp[:remaining], []byte(" "))
			if idx >= 0 {
				idx += 1
				n, err = hp.buf.Write(pp[:idx])
				hp.cursor += n
			} else {
				if len(pp)+hp.cursor > hp.writeWidth {
					if indented {
						n, err = hp.buf.Write(p[N:])
						N += n
						indented = false
						if err != nil {
							break
						}
					}
					fmt.Fprintln(hp.buf)
					hp.cursor = 0
					continue
				}
				n, err = hp.buf.Write(pp)
				hp.cursor += n
				N += n
				break
			}
		}
		indented = false
		N += n
		if err != nil {
			return N + NumIndent, err
		}
	}
	return N + NumIndent, err
}

func (hp *HelpPrinter) PrintHelp() error {
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
	err := hp.writeUsage(execStr, reqFlags, optFlags)
	if err != nil {
		return err
	}
	err = hp.writeFlagSection("Required flags", reqFlags)
	if err != nil {
		return err
	}

	err = hp.writeFlagSection("Optional flags", optFlags)
	if err != nil {
		return err
	}

	hp.buf.WriteTo(hp.out)
	return err
}
