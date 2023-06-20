package cli

import (
	"context"
)

// CommandRunner is a function that can be used to run a command.
type CommandRunner func(context.Context, interface{}) error

// Option can be used to configure a CLI.
type Option interface {
	apply(*options)
}

// ConfigObject is an option that can be used to specify a configuration object.
func ConfigObject(cfg interface{}) Option { return optionFunc(func(o *options) { o.cfgObj = cfg }) }

// Description is an option that can be used to specify a command description.
type Description struct {
	// Short is a short description of the command.
	Short string
	// Long is a long description of the command.
	Long string
}

func (d Description) apply(o *options) { o.cmdDescription = d }

// Command is an option that can be used to specify a command.
type Command struct {
	Name        string
	Description Description
	Run         CommandRunner
	Commands    []Command
}

// Commands is an option that can be used to specify multiple commands.
type Commands []Command

func (cs Commands) apply(o *options) { o.commands = append(o.commands, cs...) }

type optionFunc func(*options)

func (f optionFunc) apply(o *options) { f(o) }

type options struct {
	cfgObj         interface{}
	cmdDescription Description
	cfgFile        string

	commands []Command
}

func defaultOptions() *options { return &options{} }
