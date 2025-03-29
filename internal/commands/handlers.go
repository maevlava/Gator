package commands

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	uuid2 "github.com/google/uuid"
	"github.com/maevlava/Gator/internal/config"
	"github.com/maevlava/Gator/internal/database"
	"github.com/maevlava/Gator/internal/models"
	"html"
	"io"
	"net/http"
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

// AggHandler , the main command to return RSSFeed feed
func Agghandler(state *config.State, cmd CLI) error {
	fmt.Printf("%+v\n", fetchFeed())
	return nil
}

// handlersUtil
func checkArgsNotEmpty(cmd CLI) error {
	if len(cmd.Args) == 0 {
		return errors.New("not enough arguments")
	}
	return nil
}
func fetchFeed() *models.RSSFeed {
	// make a client with timeout context
	client := &http.Client{Timeout: time.Second * 30}

	// make a request with context
	req, err := http.NewRequestWithContext(context.Background(), "GET", "https://www.wagslane.dev/index.xml", nil)
	if err != nil {
		_ = fmt.Errorf("failed to create request: %w", err)
	}

	// do the request
	resp, err := client.Do(req)
	if err != nil {
		_ = fmt.Errorf("failed to fetch feed: %w", err)
	}
	// close the body
	defer resp.Body.Close()

	// check if status is ok
	if resp.StatusCode != http.StatusOK {
		_ = fmt.Errorf("failed to fetch feed: %s", resp.Status)
	}

	// read the body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = fmt.Errorf("failed to read body: %w", err)
	}
	// unmarshall the body to RSSFeed
	var rssFeed models.RSSFeed
	err = xml.Unmarshal(bodyBytes, &rssFeed)
	if err != nil {
		_ = fmt.Errorf("failed to unmarshal xml: %w", err)
	}

	// clean rssFeed from unescaped HTML string
	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
	for i := range rssFeed.Channel.Item {
		rssFeed.Channel.Item[i].Title = html.UnescapeString(rssFeed.Channel.Item[i].Title)
		rssFeed.Channel.Item[i].Description = html.UnescapeString(rssFeed.Channel.Item[i].Description)
	}

	return &rssFeed
}
