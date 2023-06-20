package cli

import (
	"context"
)

type CommandRunner func(context.Context, interface{}) error

type Option interface {
	apply(*options)
}

func ConfigObject(cfg interface{}) Option { return optionFunc(func(o *options) { o.cfgObj = cfg }) }

type SubCommand struct {
	Name     string
	Run      CommandRunner
	Commands []SubCommand
}

type SubCommands []SubCommand

func (scs SubCommands) apply(o *options) { o.commands = append(o.commands, scs...) }

type optionFunc func(*options)

func (f optionFunc) apply(o *options) { f(o) }

type options struct {
	cfgObj       interface{}
	cmdShortDesc string
	cmdLongDesc  string
	cfgFile      string

	commands []SubCommand
}

func defaultOptions() *options { return &options{} }
