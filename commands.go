package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/GravitiMusic/gator/internal/config"
	"github.com/GravitiMusic/gator/internal/database"
	"github.com/google/uuid"
)

type state struct {
	Database *database.Queries
	Config   *config.Config
}

type command struct {
	Name string
	Args []string
}

type commands struct {
	CommandMap map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	handler, exists := c.CommandMap[cmd.Name]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return handler(s, cmd)
}

func (c *commands) register(name string, handler func(*state, command) error) {
	c.CommandMap[name] = handler
}

func scrapeFeeds(s *state) {
	nextFeed, err := s.Database.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Printf("failed to get next feed to fetch")
		return
	}

	fmt.Println("Found a feed to fetch:", nextFeed.Name)
	scrapeFeed(s, nextFeed)
}

func scrapeFeed(s *state, feed database.Feed) {
	err := s.Database.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		fmt.Printf("failed to mark feed as fetched: %v\n", err)
		return
	}

	feedData, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Printf("failed to fetch feed data: %v\n", err)
		return
	}

	for _, item := range feedData.Channel.Item {
		formats := []string{time.RFC1123Z, time.RFC1123, time.RFC3339}
		var pubTime time.Time
		for _, format := range formats {
			t, err := time.Parse(format, item.PubDate)
			if err == nil {
				pubTime = t
				break
			}
		}

		_, err := s.Database.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			PublishedAt: sql.NullTime{Time: pubTime, Valid: !pubTime.IsZero()},
			FeedID: feed.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				continue
			}
			fmt.Printf("failed to create post: %v\n", err)
			continue
		}


	}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("login command requires an argument")
	}

	if _, err := s.Database.GetUser(context.Background(), cmd.Args[0]); err != nil {
		fmt.Printf("user '%s' does not exist: %v\n", cmd.Args[0], err)
		os.Exit(1)
	}

	err := s.Config.SetUser(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("failed to set user: %w", err)
	}

	fmt.Println("User has been set to:", cmd.Args[0])
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("register command requires an argument")
	}

	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
	}

	if _, err := s.Database.GetUser(context.Background(), cmd.Args[0]); err == nil {
		fmt.Printf("user '%s' already exists\n", cmd.Args[0])
		os.Exit(1)
	}

	user, err := s.Database.CreateUser(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	err = s.Config.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("failed to set user: %w", err)
	}

	fmt.Printf("User '%s' has been registered and set as the current user.\n", user.Name)
	return nil
}

func handlerReset(s *state, cmd command) error {
	if err := s.Database.DeleteAllUsers(context.Background()); err != nil {
		return fmt.Errorf("failed to reset users: %w", err)
	}
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.Database.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}

	for _, user := range users {
		if user.Name == s.Config.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("agg command requires exactly one argument")
	}

	time_between_reqs, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("failed to parse duration: %w", err)
	}
	
	fmt.Printf("Collecting feeds every %s...\n", time_between_reqs)

	ticker := time.NewTicker(time_between_reqs)

	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerAddfeed(s *state, cmd command, user database.User) error {
	currentName := s.Config.CurrentUserName
	if currentName == "" {
		return fmt.Errorf("no user is currently logged in")
	}

	if len(cmd.Args) != 2 {
		os.Exit(1)
	}

	name := cmd.Args[0]
	url := cmd.Args[1]

	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}

	feed, err := s.Database.CreateFeed(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to create feed: %w", err)
	}

	feedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	if _, err := s.Database.CreateFeedFollow(context.Background(), feedFollowParams); err != nil {
		return fmt.Errorf("failed to create feed follow: %w", err)
	}

	fmt.Printf("Feed fields: \nID: %s\nCreatedAt: %s\nUpdatedAt: %s\nName: %s\nUrl: %s\nUserID: %s\n", feed.ID, feed.CreatedAt, feed.UpdatedAt, feed.Name, feed.Url, feed.UserID)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.Database.GetAllFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get feeds: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("Name: %s\nUrl: %s\nUser: %s\n\n", feed.Name, feed.Url, feed.Name_2)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("follow command requires exactly one argument")
	}

	currentName := s.Config.CurrentUserName
	if currentName == "" {
		return fmt.Errorf("no user is currently logged in")
	}

	url := cmd.Args[0]
	feed, err := s.Database.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("failed to get feed by URL: %w", err)
	}

	feedID := feed.ID
	userID := user.ID

	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    userID,
		FeedID:    feedID,
	}

	feedFollow, err := s.Database.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to create feed follow: %w", err)
	}

	fmt.Printf("Feed %s is now followed by user %s\n", feedFollow.FeedName, feedFollow.UserName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	currentUserName := s.Config.CurrentUserName
	if currentUserName == "" {
		return fmt.Errorf("no user is currently logged in")
	}

	follow, err := s.Database.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to get feed follows for user: %w", err)
	}

	for _, f := range follow {
		fmt.Printf("Feed: %s\n", f.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("unfollow command requires exactly one argument")
	}

	currentName := s.Config.CurrentUserName
	if currentName == "" {
		return fmt.Errorf("no user is currently logged in")
	}

	url := cmd.Args[0]
	feed, err := s.Database.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("failed to get feed by URL: %w", err)
	}

	params := database.DeleteFeedFollowByUserAndFeedParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	if err := s.Database.DeleteFeedFollowByUserAndFeed(context.Background(), params); err != nil {
		return fmt.Errorf("failed to unfollow feed: %w", err)
	}

	fmt.Printf("Feed %s is no longer followed by user %s\n", feed.Name, user.Name)
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	var limit int = 2
	if len(cmd.Args) == 1{
		limit, _ = strconv.Atoi(cmd.Args[0])
	}

	posts, err := s.Database.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: uuid.MustParse(user.ID.String()),
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("failed to get posts for user: %w", err)
	}

	for _, post := range posts {
		fmt.Printf("Title: %s\nUrl: %s\nDescription: %s\nPublishedAt: %s\n\n", post.Title, post.Url, post.Description.String, post.PublishedAt.Time)
	}
	return nil
}