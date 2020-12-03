package commander

import (
	"io"

	kingpin "github.com/tj/kingpin"
)

// New commander
func New(name, help string) *Command {
	root := kingpin.New(name, help).Terminate(nil)

	// support -h
	root.HelpFlag.Short('h')

	return &Command{root: root}
}

// Command struct
type Command struct {
	root *kingpin.Application
}

// Version sets the version
func (c *Command) Version(version string) {
	c.root.Version(version)
}

// Command creates a command
func (c *Command) Command(name, help string) *Subcommand {
	cmd := c.root.Command(name, help)
	return &Subcommand{
		root: c,
		cmd:  cmd,
	}
}

// Writer adjusts where the command is written out. The default is os.Stderr.
func (c *Command) Writer(w io.Writer) *Command {
	c.root.UsageWriter(w)
	return c
}

// Flag adds a command flag
func (c *Command) Flag(name, help string) *kingpin.FlagClause {
	return c.root.Flag(name, help)
}

// Arg adds a command argument
func (c *Command) Arg(name, help string) *kingpin.ArgClause {
	return c.root.Arg(name, help)
}

// Example adds an example
func (c *Command) Example(usage, help string) {
	c.root.Example(usage, help)
}

// Before runs a command before run
func (c *Command) Before(fn func() error) {
	c.root.PreAction(func(_ *kingpin.ParseContext) error {
		return fn()
	})
}

// Run doesn't do anything on the root
func (c *Command) Run(fn func() error) {
	c.root.Action(func(_ *kingpin.ParseContext) error {
		return fn()
	})
}

// Parse the args
func (c *Command) Parse(args []string) error {
	_, err := c.root.Parse(args)
	return err
}

// MustParse the args
func (c *Command) MustParse(args []string) {
	if err := c.Parse(args); err != nil {
		c.Fatal(err)
	}
}

// Fatal error
func (c *Command) Fatal(err error) {
	c.root.FatalUsage(err.Error())
}

// Usage displays the help
func (c *Command) Usage() {
	c.root.Usage([]string{})
}

// Subcommand struct
type Subcommand struct {
	root *Command
	cmd  *kingpin.Cmd
}

// Command creates a subcommand
func (c *Subcommand) Command(name, help string) *Subcommand {
	sub := c.cmd.Command(name, help)
	return &Subcommand{
		root: c.root,
		cmd:  sub,
	}
}

// Flag adds a command flag
func (c *Subcommand) Flag(name, help string) *kingpin.FlagClause {
	return c.cmd.Flag(name, help)
}

// Arg adds a command argument
func (c *Subcommand) Arg(name, help string) *kingpin.ArgClause {
	return c.cmd.Arg(name, help)
}

// Default makes the command the default
func (c *Subcommand) Default() *Subcommand {
	c.cmd.Default()
	return c
}

// Alias adds an alias for this command
func (c *Subcommand) Alias(name string) *Subcommand {
	c.cmd.Alias(name)
	return c
}

// Use fn
func (c *Subcommand) Use(fn func(c *Subcommand) error) {
	fn(c)
}

// Example adds an example
func (c *Subcommand) Example(usage, help string) {
	c.cmd.Example(usage, help)
}

// Before runs a command before run
func (c *Subcommand) Before(fn func() error) {
	c.cmd.PreAction(func(_ *kingpin.ParseContext) error {
		return fn()
	})
}

// Run executes if this command is run
func (c *Subcommand) Run(fn func() error) {
	c.cmd.Action(func(_ *kingpin.ParseContext) error {
		return fn()
	})
}

// Parse doesn't do anything on a command
func (c Subcommand) Parse(args []string) error {
	return nil
}
