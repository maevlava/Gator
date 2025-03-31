package commands

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	uuid2 "github.com/google/uuid"
	"github.com/maevlava/Gator/internal/config"
	"github.com/maevlava/Gator/internal/database"
	"github.com/maevlava/Gator/internal/models"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Middleware Logged In Handlers

func MiddlewareLoggedIn(handler func(state *config.State, cmd CLI, user database.User) error) func(state *config.State, cmd CLI) error {

	return func(state *config.State, cmd CLI) error {
		user, err := state.DB.GetUser(context.Background(), state.Config.CurrentUser)
		if err != nil {
			return fmt.Errorf("failed to get current user '%v': %w", state.Config.CurrentUser, err)
		}
		err = handler(state, cmd, user)

		return err
	}
}

// AddFeedHandler, To create feed by a current user
func AddFeedHandler(state *config.State, cmd CLI, currentUser database.User) error {
	// need 2 args
	if len(cmd.Args) < 2 {
		return errors.New("not enough arguments: feedName and feedUrl are required")
	}
	feedName := cmd.Args[0]
	feedUrl := cmd.Args[1]

	createTime := time.Now().UTC()
	createFeedParams := database.CreateFeedParams{
		ID:        uuid2.New(),
		CreatedAt: createTime,
		UpdatedAt: createTime,
		Name:      feedName,
		Url:       feedUrl,
		UserID:    currentUser.ID,
	}

	feed, err := state.DB.CreateFeed(context.Background(), createFeedParams)
	if err != nil {
		return fmt.Errorf("failed to create feed '%s': %w", feedName, err)
	}

	// enhanced to auto current user follow the feed
	followTime := time.Now().UTC()
	createFollowParams := database.CreateFeedFollowParams{
		ID:        uuid2.New(),
		CreatedAt: followTime,
		UpdatedAt: followTime,
		UserID:    currentUser.ID,
		FeedID:    feed.ID,
	}

	_, err = state.DB.CreateFeedFollow(context.Background(), createFollowParams)
	if err != nil {
		return err
	}

	return nil
}

// FollowHandler to create new Feed Follow for current user
func FollowHandler(state *config.State, cmd CLI, currentUser database.User) error {

	// need 1 arg
	var err error = nil

	if len(cmd.Args) < 1 {
		_ = fmt.Errorf("not enough arguments: feedUrl is required")
		os.Exit(1)
	}

	feedUrl := cmd.Args[0]

	feed, err := state.DB.GetFeedByUrl(context.Background(), feedUrl)
	if err != nil {
		return errors.New("failed to get feed by url")
	}

	now := time.Now()
	feedFollowsParams := database.CreateFeedFollowParams{
		ID:        uuid2.New(),
		CreatedAt: now,
		UpdatedAt: now,
		FeedID:    feed.ID,
		UserID:    currentUser.ID,
	}
	newFeedFollows, err := state.DB.CreateFeedFollow(context.Background(), feedFollowsParams)
	if err != nil {
		return fmt.Errorf("failed to create feed follow for feed '%s' by user '%s': %w", feed.Name, currentUser.Name, err)
	}

	fmt.Println("FeedFollows created successfully")
	fmt.Printf("Feed: %v\n", newFeedFollows)

	return err
}

// FollowingHandler return all the feeds current user are following
func FollowingHandler(state *config.State, cmd CLI, currentUser database.User) error {
	followedFeeds, err := state.DB.GetFollowedFeedsForUser(context.Background(), currentUser.ID)
	if err != nil {
		return fmt.Errorf("failed to get followed feeds for user '%s': %w", currentUser.Name, err)
	}

	for _, feed := range followedFeeds {
		fmt.Printf("- %s\n", feed.Name)
	}
	fmt.Printf("%s\n", currentUser.Name)

	return nil // Success
}

// UnfollowHandler delete a feed follow and feed follow combination
func UnfollowHandler(state *config.State, cmd CLI, currentUser database.User) error {

	if len(cmd.Args) < 1 {
		return errors.New("not enough arguments: feedUrl is required")
	}
	feedUrl := cmd.Args[0]

	DeleteFeedFollowParams := database.DeleteFeedFollowForUserParams{
		UserID: currentUser.ID,
		Url:    feedUrl,
	}
	err := state.DB.DeleteFeedFollowForUser(context.Background(), DeleteFeedFollowParams)
	if err != nil {
		return fmt.Errorf("failed to delete feed follow for user '%s': %w", currentUser.Name, err)
	}

	fmt.Printf("Successfully unfollowed feed: %s\n", feedUrl)

	return nil
}

func BrowseHandler(state *config.State, cmd CLI, user database.User) error {
	limit := int32(2)
	if len(cmd.Args) > 0 {
		parsedLimit, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return err
		}
		if parsedLimit <= 0 {
			return errors.New("limit must be a positive integer")
		}
		limit = int32(parsedLimit)
	}

	posts, err := state.DB.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  limit,
	})
	if err != nil {
		return fmt.Errorf("failed to get posts for user '%s': %w", user.Name, err)
	}

	for _, post := range posts {
		fmt.Printf("Title: %s\n", post.Title)

		publishedStr := "N/A"
		if post.PublishedAt.Valid {
			publishedStr = post.PublishedAt.Time.Format(time.RFC1123)
		}
		fmt.Printf("Published: %s\n", publishedStr)
		descriptionStr := "No description available."
		if post.Description.Valid && post.Description.String != "" {
			descriptionStr = post.Description.String
		}

		fmt.Printf("Description: %s\n", descriptionStr)
		fmt.Printf("URL: %s\n", post.Url)
	}

	return nil
}

// Basic Handler

// LoginHandler set current_user to user login
func LoginHandler(state *config.State, cmd CLI) error {

	// handle 0 argument
	if len(cmd.Args) == 0 {
		return errors.New("not enough arguments")
	}

	// set current user to logged user
	_, err := state.DB.GetUser(context.Background(), cmd.Args[0])
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
	if len(cmd.Args) == 0 {
		return errors.New("not enough arguments")
	}
	name := cmd.Args[0]

	// Check if the user already exists
	_, err := state.DB.GetUser(context.Background(), name)
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
			continue
		}
		fmt.Printf("* %s\n", user.Name)
	}

	return err
}

// FeedListHandler, to print out all name, url, and creator  of the feed
func FeedListHandler(state *config.State, cmd CLI) error {
	feedListWithUser, err := state.DB.GetAllFeedsWithUser(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get all feeds: %w", err)
	}

	for _, feed := range feedListWithUser {
		fmt.Printf("%s\n", feed.Name)
		fmt.Printf("%s\n", feed.Url)
		fmt.Printf("%s\n", feed.UserName)
	}

	return err
}

// --- Aggregator Code ---
// fetchAndParseFeed Takes a URL and returns the parsed feed or an error.
func fetchAndParseFeed(url string) (*models.RSSFeed, error) {
	client := &http.Client{Timeout: time.Second * 30}

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", url, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch feed %s: status code %d", url, resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body for %s: %w", url, err)
	}

	var rssFeed models.RSSFeed
	err = xml.Unmarshal(bodyBytes, &rssFeed)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal xml for %s: %w", url, err)
	}

	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
	for i := range rssFeed.Channel.Item {
		rssFeed.Channel.Item[i].Title = html.UnescapeString(rssFeed.Channel.Item[i].Title)
		rssFeed.Channel.Item[i].Description = html.UnescapeString(rssFeed.Channel.Item[i].Description)
	}

	return &rssFeed, nil
}

// scrapeFeeds
func scrapeFeeds(db *database.Queries) error {
	feed, err := db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("db.GetNextFeedToFetch: %w", err)
	}

	now := time.Now().UTC()
	markParams := database.MarkFeedFetchedParams{
		ID:            feed.ID,
		LastFetchedAt: sql.NullTime{Time: now, Valid: true},
	}
	err = db.MarkFeedFetched(context.Background(), markParams)
	if err != nil {
		return fmt.Errorf("db.MarkFeedFetched feed ID %s: %w", feed.ID, err)
	}

	parsedFeed, err := fetchAndParseFeed(feed.Url)
	if err != nil {
		return fmt.Errorf("fetchAndParseFeedA %s: %w", feed.Url, err)
	}

	processedCount := 0
	skippedCount := 0
	for _, item := range parsedFeed.Channel.Item {

		publishedAt := sql.NullTime{}
		dateString := item.PubDate

		if dateString != "" {
			parsedTime, err := time.Parse(time.RFC1123Z, dateString)
			if err != nil {
				parsedTime, err = time.Parse(time.RFC1123, dateString)
			}

			if err == nil {
				publishedAt = sql.NullTime{Time: parsedTime.UTC(), Valid: true}
			} else {
				return fmt.Errorf("failed to parse date %s: %w", dateString, err)
			}
		}

		description := sql.NullString{}
		if item.Description != "" {
			description = sql.NullString{String: item.Description, Valid: true}
		}

		postUrl := item.Link
		if postUrl == "" {
			log.Printf("Skipping post '%s' - missing URL", item.Title)
			skippedCount++
			continue
		}

		createParams := database.CreatePostParams{
			ID:          uuid2.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Url:         postUrl,
			Description: description,
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		}

		_, err = db.CreatePost(context.Background(), createParams)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				skippedCount++
				continue
			}
			log.Printf("Failed to create post '%s' (%s): %v", item.Title, postUrl, err)
			continue
		}
		processedCount++
		fmt.Printf("   - Post Saved: %s\n", item.Title) // Kept this print for user feedback
	}
	return nil
}

// AggHandler runs the main feed aggregation loop.
func AggHandler(state *config.State, cmd CLI) error {
	if len(cmd.Args) < 1 {
		return errors.New("time_between_reqs argument required (e.g., '1m', '30s')")
	}
	durationStr := cmd.Args[0]
	timeBetweenRequests, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("invalid duration format '%s': %w", durationStr, err)
	}

	minDuration := 5 * time.Second
	if timeBetweenRequests < minDuration {
		log.Printf("Duration is too short")
	}

	ticker := time.NewTicker(timeBetweenRequests)
	defer ticker.Stop()

	if err := scrapeFeeds(state.DB); err != nil {
		log.Printf("Initial scrape failed: %v", err)
	}

	for range ticker.C {
		if err := scrapeFeeds(state.DB); err != nil {
			log.Printf("Scraping failed during loop: %v", err)
		}
	}

	return nil
}
