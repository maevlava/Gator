package commands

import (
	"fmt"
	"github.com/maevlava/Gator/internal/config"
)

type CLI struct {
	Name string
	Args []string
}
type Registry struct {
	Commands map[string]func(state *config.State, command CLI) error
}

func (c *Registry) Register(name string, f func(*config.State, CLI) error) {
	// register command
	c.Commands[name] = f
}
func (c *Registry) Run(s *config.State, cmd CLI) error {
	// find command
	command, ok := c.Commands[cmd.Name]

	if !ok {
		return fmt.Errorf("command %s not found", cmd.Name)
	}

	// if exist, execute it with current state e.g. user
	err := command(s, cmd)
	if err != nil {
		return err
	}

	return nil
}
