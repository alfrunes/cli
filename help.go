package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

const (
	defaultWidth int = 80
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

func writeFlagSection(w io.Writer, section string, flags []Flag) error {
	_, err := w.Write([]byte(section + ":\n"))
	if err != nil {
		return err
	}
	return nil
}

func writeUsage(
	w io.Writer, width int,
	execStr string,
	required, optional []*Flag,
) error {
	var flagWords []string
	var idx int
	var indent string
	var word string = fmt.Sprintf("Usage: %s ", execStr)

	n, err := w.Write([]byte(word))
	if err != nil {
		return err
	}
	idx += n
	if n < width {
		indent = fmt.Sprintf("%*s", n, " ")
	}

	flagWords = make([]string, len(optional)+len(required))
	for i, flag := range required {
		word := "--" + flag.Name
		if flag.Char != rune(0) {
			word += "/-" + string(flag.Char)
		}
		if flag.Type != Bool && flag.MetaVar == "" {
			word += " value] "
		} else {
			word += flag.MetaVar + " "
		}
		flagWords[i] = word
	}
	for i, flag := range optional {
		word := "[--" + flag.Name
		if flag.Char != rune(0) {
			word += "/-" + string(flag.Char)
		}
		if flag.Type == Bool {
			word += "] "
		} else if flag.MetaVar == "" {
			word += " value] "
		} else {
			word += fmt.Sprintf(" %s] ", flag.MetaVar)
		}
		flagWords[i+len(required)] = word
	}

	for i := 0; i < len(flagWords); i++ {
		word = flagWords[i]
		if idx+len(word) > width {
			word = NewLine + indent + word
			n, err = w.Write([]byte(word))
			idx = n
		} else {
			n, err = w.Write([]byte(word))
			idx += n
		}
		if err != nil {
			return err
		}
	}
	_, err = w.Write([]byte(NewLine))

	return err
}

func PrintAppHelp(app *App, out io.Writer) error {
	var width int
	helpBuf := &bytes.Buffer{}
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

	optFlags, reqFlags := getOptionalAndRequired(app.Flags)
	err := writeUsage(helpBuf, width, os.Args[0], reqFlags, optFlags)

	helpBuf.WriteTo(os.Stderr)
	return err
}
