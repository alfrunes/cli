package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Flag interface {
	String() string
	GetValue() interface{}
	Validate() error
	Set(string) error
	GetProperties() flagProperties
}

type flagProperties struct {
	Name     string
	Char     rune
	EnvVar   string
	Required bool
	IsSet    bool
}

type StringFlag struct {
	// Name of the flag, for a given Name the command-line option
	// becomes --Name.
	Name string
	// Char is an optional single-char alternative
	Char rune
	// Initialize default value from environment variable.
	EnvVar string
	// Required makes the flag required.
	Required bool
	// Usage is printed to the help screen - short summary of function.
	Usage string
	// Value holds the default (string) value of the flag (defaults to "").
	Value string
	// Choices restricts the Values this flag can take to this set.
	Choices []string

	isInitialized bool
}

func (f *StringFlag) Set(value string) error {
	f.Value = value
	return f.Validate()
}

func (f *StringFlag) String() string {
	usage := f.Usage
	if len(f.Choices) != 0 {
		usage += fmt.Sprintf(" {%s}", strings.Join(f.Choices, ", "))
	}
	if f.Value != "" {
		usage += fmt.Sprintf(" [%s]", f.Value)
	}
	return f.Usage
}

func (f *StringFlag) GetProperties() flagProperties {
	return flagProperties{
		Name:     f.Name,
		Char:     f.Char,
		EnvVar:   f.EnvVar,
		Required: f.Required,
	}
}

func (f *StringFlag) GetValue() interface{} {
	return interface{}(f.Value)
}

func (f *StringFlag) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("StringFlag is missing name")
	}
	if !f.isInitialized && f.EnvVar != "" {
		env := os.Getenv(f.EnvVar)
		if env != "" {
			f.Value = env
		}
		f.isInitialized = true
	}
	if len(f.Choices) != 0 {
		for _, v := range f.Choices {
			if f.Value == v {
				return nil
			}
		}
		return fmt.Errorf(
			"illegal value for string flag '%s': %s not in {%s}",
			f.Name, f.Value, strings.Join(f.Choices, ", "))
	}
	return nil
}

type IntFlag struct {
	// Name of the flag, for a given Name the command-line option
	// becomes --Name.
	Name string
	// Char is an optional single-char alternative
	Char rune
	// Initialize default value from environment variable.
	EnvVar string
	// Required makes the flag required.
	Required bool
	// Usage is printed to the help screen - short summary of function.
	Usage string
	// Value holds the default (integer) value of the flag (defaults to 0).
	Value int
	// Range restricts the range of the flag to the selected values.
	Range [2]int
}

func (f *IntFlag) GetValue() interface{} {
	return interface{}(f.Value)
}

func (f *IntFlag) GetProperties() flagProperties {
	return flagProperties{
		Name:     f.Name,
		Char:     f.Char,
		EnvVar:   f.EnvVar,
		Required: f.Required,
	}
}

func (f *IntFlag) Set(value string) error {
	var err error
	f.Value, err = strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid value for integer flag %s: %s",
			f.Name, value)
	}
	return f.Validate()
}

func (f *IntFlag) String() string {
	var hasRange bool = false
	usage := f.Usage
	if f.Range[0] != f.Range[1] {
		usage += fmt.Sprintf(" {%d-%d}", f.Range[0], f.Range[1])

	}
	if f.Value != 0 || hasRange {
		usage += fmt.Sprintf(" [%s]", f.Value)
	}
	return f.Usage
}

func (f *IntFlag) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("IntFlag is missing name")
	}
	if f.Value < f.Range[0] {
		return fmt.Errorf("illegal value for integer flag %s: %d < %d",
			f.Name, f.Value, f.Range[0])
	} else if f.Value > f.Range[1] {
		return fmt.Errorf("illegal value for integer flag %s: %d > %d",
			f.Name, f.Value, f.Range[1])
	}

	return nil
}

type BoolFlag struct {
	// Name of the flag, for a given Name the command-line option
	// becomes --Name.
	Name string
	// Char is an optional single-char alternative
	Char rune
	// Initialize default value from environment variable.
	EnvVar string
	// Required makes the flag required.
	Required bool
	// Usage is printed to the help screen - short summary of function.
	Usage string
	// Value is the default (boolean) value of the flag (defaults to false).
	Value bool
	// PrintDefault determines if the Stringer is printing the default value.
	PrintDefault bool
}

func (f *BoolFlag) GetProperties() flagProperties {
	return flagProperties{
		Name:     f.Name,
		Char:     f.Char,
		EnvVar:   f.EnvVar,
		Required: f.Required,
	}
}

func (f *BoolFlag) GetValue() interface{} {
	return interface{}(f.Value)
}

func (f *BoolFlag) Set(value string) error {
	lowerCase := strings.ToLower(value)
	if lowerCase == "true" {
		f.Value = true
		return nil
	} else if lowerCase == "false" {
		f.Value = false
		return nil
	}
	return fmt.Errorf("illegal value: %s", value)
}

// Prints the usage string of the flag.
func (f *BoolFlag) String() string {
	if f.PrintDefault {
		return fmt.Sprintf("%s [%s]", f.Usage, f.Value)
	}
	return f.Usage
}
func (f *BoolFlag) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("BoolFlag is missing name")
	}
	return nil
}
