package commands

import (
	"errors"
	"fmt"
	"github.com/maevlava/Gator/internal/config"
)

// LoginHandler set current_user to user login
func LoginHandler(state *config.State, cmd CLI) error {

	// handle 0 argument
	if len(cmd.Args) == 0 {
		return errors.New("not enough arguments")
	}

	// set current user to logged user
	err := state.Config.SetUser(cmd.Args[0])
	if err != nil {
		return errors.New("failed to set user")
	}

	// print message to terminal user has been set
	fmt.Println("User has been set")
	return nil
}
