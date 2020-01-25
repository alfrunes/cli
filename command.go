package cli

// Command describes git-style commands such as `git <log|diff|commit>` etc.
// Each Command has it's own scope of flags and possible SubCommands.
type Command struct {
	// Name of the command.
	Name string

	// Action is the bootstrapping function of the command.
	Action func(*Context) error

	// Description contains a *longer* description of the command.
	Description string
	// Usage should give a short summary of the description.
	Usage string

	// Flags that the command accepts.
	Flags []*Flag
	// InheritParentFlags toggles whether the flags of the parent command (or
	// app) is accessible at the command's scope.
	InheritParentFlags bool
	// SubCommands are commands that are accessible under this scope.
	SubCommands []*Command
}

func (cmd *Command) PrintHelp() {
	return // TODO
}

func (cmd *Command) Validate() error {
	return nil // TODO
}
