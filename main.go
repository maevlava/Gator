package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/maevlava/Gator/internal/commands"
	"github.com/maevlava/Gator/internal/config"
	"github.com/maevlava/Gator/internal/database"
	"os"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		_ = fmt.Errorf("error reading config file: %v", err)
	}
	state := &config.State{Config: &cfg}
	commandsRegistry := &commands.Registry{
		Commands: make(map[string]func(state *config.State, command commands.CLI) error),
	}
	db, err := sql.Open("postgres", state.Config.DBUrl)
	state.DB = database.New(db)

	registerCommandsHandlers(commandsRegistry)
	listenToCommands(state, commandsRegistry)

}

func listenToCommands(state *config.State, commandsRegistry *commands.Registry) {
	args := os.Args
	if len(args) < 3 {
		_ = fmt.Errorf("not enough arguments")
		os.Exit(1)
	}
	command := commands.CLI{Name: args[1], Args: args[2:]}
	err := commandsRegistry.Run(state, command)
	if err != nil {
		_ = fmt.Errorf("error running command: %s\n%v", command.Name, err)
	}
}

func registerCommandsHandlers(commandsRegistry *commands.Registry) {
	commandsRegistry.Register("login", commands.LoginHandler)

}
