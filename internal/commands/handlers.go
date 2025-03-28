package commands

import (
	"context"
	"errors"
	"fmt"
	uuid2 "github.com/google/uuid"
	"github.com/maevlava/Gator/internal/config"
	"github.com/maevlava/Gator/internal/database"
	"os"
	"time"
)

// LoginHandler set current_user to user login
func LoginHandler(state *config.State, cmd CLI) error {

	// handle 0 argument
	err := checkArgsNotEmpty(cmd)
	if err != nil {
		return err
	}

	// set current user to logged user
	_, err = state.DB.GetUser(context.Background(), cmd.Args[0])
	if err != nil {
		fmt.Println("user does not exist")
		os.Exit(1)
	}
	err = state.Config.SetUser(cmd.Args[0])
	if err != nil {
		return errors.New("failed to set user")
	}

	// print message to terminal user has been set
	fmt.Println("User has been set")
	return nil
}

// RegisterHandler register user to table users in postgres database
func RegisterHandler(state *config.State, cmd CLI) error {
	// args for user's name
	err := checkArgsNotEmpty(cmd)
	if err != nil {
		return err
	}
	name := cmd.Args[0]

	// Check if the user already exists
	_, err = state.DB.GetUser(context.Background(), name)
	if err == nil {
		// User exists exit code 1
		fmt.Println("User already exists")
		os.Exit(1)
	}

	// generate uuid, created_at, updated at
	id := uuid2.New()
	now := time.Now()
	createUserParams := database.CreateUserParams{
		ID:        id,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// register the name using db.queries
	user, err := state.DB.CreateUser(context.Background(), createUserParams)
	if err != nil {
		return err
	}

	// set current user
	state.Config.SetUser(name)

	// Print success message and user data
	fmt.Println("User created successfully")
	fmt.Printf("User: %v\n", user)
	return nil
}

// ResetHandler reset the users table (delete all the rows)
func ResetHandler(state *config.State, cmd CLI) error {
	// tell state to delete all the rows
	err := state.DB.DeleteAllUser(context.Background())
	if err != nil {
		return fmt.Errorf("failed to delete all users: %w", err)
	}
	return err
}

// UserListHandler list all users in the users database
func UserListHandler(state *config.State, cmd CLI) error {
	users, err := state.DB.GetAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get all users: %w", err)
	}
	for _, user := range users {
		if user.Name == state.Config.CurrentUser {
			fmt.Printf("* %s (current)\n", user.Name)
		}
		fmt.Printf("* %s", user.Name)
	}

	return err
}

// handlersUtil
func checkArgsNotEmpty(cmd CLI) error {
	if len(cmd.Args) == 0 {
		return errors.New("not enough arguments")
	}
	return nil
}
