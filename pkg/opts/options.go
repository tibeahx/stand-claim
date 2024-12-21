package opts

import "gopkg.in/telebot.v4"

type Options struct {
	stands   []string
	commands []telebot.Command
}

var defaultStands = []string{
	"dev1",
	"dev2",
	"dev3",
	"dev4",
}

func (o Options) Stands() []string {
	if o.stands == nil {
		return defaultStands
	}
	return o.stands
}

var defaultCommands = []telebot.Command{
	{Text: "claim", Description: "Claim a stand"},
	{Text: "release", Description: "Release currently claimed stand"},
	{Text: "status", Description: "Show current stand status"},
	{Text: "list", Description: "Show all stands"},
	{Text: "list_free", Description: "Show available stands"},
	{Text: "ping", Description: "Ping current stand owner"},
}

func (o Options) Commands() []telebot.Command {
	if o.commands == nil {
		return defaultCommands
	}
	return o.commands
}
