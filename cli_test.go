package cli

import "fmt"

func ExampleApp() {
	// Getting Started with cli:
	// There are only two steps for using this package:
	// 1 - Define your App structure
	// 2 - Run your app by passing commandline arguments to App.Run()

	exampleAction := func(ctx *Context) error {
		fmt.Println("Hello World")
		// ctx.String/ctx.Bool/ctx.Int/ctx.Float returns the flag value
		// and whether the flag was explicitly set.
		val, set := ctx.String("example-boi")
		fmt.Printf("This is my flag value: %s\n", val)
		fmt.Printf("Was this flag set? %v\n", set)

		fmt.Println()
		fmt.Println("This is the main help text:")
		fmt.Println("```")
		ctx.GetParent().PrintHelp()

		fmt.Println("```")
		fmt.Println("Where as this is the usage text:")
		fmt.Println("````")
		ctx.GetParent().PrintUsage()
		fmt.Println("```")

		return nil
	}

	app := App{
		Name:        "example",
		Description: "Describe your app here...",
		Flags: []*Flag{
			{
				Name: "example-boi",
				Char: 'e',
				// String is default type if none is specified.
				Type:     String,
				MetaVar:  "STR",
				Default:  "default value",
				Choices:  []string{"must", "include", "default value"},
				EnvVar:   "INIT_FROM_ENVIRONMENT_VAR_IF_DEFINED",
				Required: false, // false is default
				Usage:    "Doesn't do much...",
			},
		},
		Commands: []*Command{
			{
				Name:               "example-cmd",
				Action:             exampleAction,
				Description:        "Describe me here...",
				Usage:              "Short summary of Description",
				InheritParentFlags: false,
				PositionalArguments: []string{"these", "will",
					"appear", "in", "usage", "text"},
				SubCommands: nil,
			},
		},
	}
	app.Run([]string{"example", "-e", "include", "example-cmd"})
	// Output:
	// Hello World
	// This is my flag value: include
	// Was this flag set? true
	//
	// This is the main help text:
	// ```
	// Usage: example [-e STR] [-h] {example-cmd,help}
	//
	// Description:
	//   Describe your app here...
	//
	// Commands:
	//   example-cmd           Short summary of Description
	//   help                  Show help for command given as argument
	//
	// Optional flags:
	//   --example-boi/-e STR  Doesn't do much... {must, include, default value}
	//   --help/-h             Display this help message
	// ```
	// Where as this is the usage text:
	// ````
	// Usage: example [-e STR] [-h] {example-cmd,help}
	// ```
}
