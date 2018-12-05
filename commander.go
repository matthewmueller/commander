package commander

import (
	kingpin "github.com/tj/kingpin"
)

// // Commander of commander
// type Commander interface {
// 	Command(name, help string) Commander
// 	Use(fn func(child Commander) error)
// 	Flag(name, help string) *kingpin.FlagClause
// 	Arg(name, help string) *kingpin.ArgClause
// 	Run(fn func() error)
// 	Parse(args []string) error
// 	Default() Commander
// }

// Root of the CLI
type Root struct {
	root *kingpin.Application
}

// New commander
func New(name, help string) *Root {
	root := kingpin.New(name, help)

	// support -h
	root.HelpFlag.Short('h')

	return &Root{root: root}
}

// Version sets the version
func (c *Root) Version(version string) {
	c.root.Version(version)
}

// Command creates a command
func (c *Root) Command(name, help string) *Command {
	cmd := c.root.Command(name, help)
	return &Command{
		root: c,
		cmd:  cmd,
	}
}

// Flag adds a command flag
func (c *Root) Flag(name, help string) *kingpin.FlagClause {
	return c.root.Flag(name, help)
}

// Arg adds a command argument
func (c *Root) Arg(name, help string) *kingpin.ArgClause {
	return c.root.Arg(name, help)
}

// Default doesn't do anything on the root
func (c *Root) Default() *Root {
	return c
}

// Use fn
func (c *Root) Use(fn func(c *Root) error) {
	fn(c)
}

// Run doesn't do anything on the root
func (c *Root) Run(fn func() error) {
	c.root.Action(func(_ *kingpin.ParseContext) error {
		return fn()
	})
}

// Parse the args
func (c *Root) Parse(args []string) error {
	_, err := c.root.Parse(args)
	return err
}

// Fatal error
func (c *Root) Fatal(err error) {
	c.root.FatalUsage(err.Error())
}

// Usage displays the help
func (c *Root) Usage() {
	c.root.Usage([]string{})
}

// Command struct
type Command struct {
	root *Root
	cmd  *kingpin.Cmd
}

// Command creates a subcommand
func (c *Command) Command(name, help string) *Command {
	sub := c.cmd.Command(name, help)
	return &Command{
		root: c.root,
		cmd:  sub,
	}
}

// Flag adds a command flag
func (c *Command) Flag(name, help string) *kingpin.FlagClause {
	return c.cmd.Flag(name, help)
}

// Arg adds a command argument
func (c *Command) Arg(name, help string) *kingpin.ArgClause {
	return c.cmd.Arg(name, help)
}

// Default makes the command the default
func (c *Command) Default() *Command {
	c.cmd.Default()
	return c
}

// Alias adds an alias for this command
func (c *Command) Alias(name string) *Command {
	c.cmd.Alias(name)
	return c
}

// Use fn
func (c *Command) Use(fn func(c *Command) error) {
	fn(c)
}

// Run executes if this command is run
func (c *Command) Run(fn func() error) {
	c.cmd.Action(func(_ *kingpin.ParseContext) error {
		return fn()
	})
}

// Parse doesn't do anything on a command
func (c Command) Parse(args []string) error {
	return nil
}
