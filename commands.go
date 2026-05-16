package main

import (
	"context"
	"fmt"
	"os"
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

func handlerResest(s *state, cmd command) error {
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