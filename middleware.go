package main

import (
	"context"
	"fmt"

	"github.com/GravitiMusic/gator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(s *state, cmd command) error {
	return func(s *state, cmd command) error {
		if s.Config.CurrentUserName == "" {
			return fmt.Errorf("no user is currently logged in")
		}

		user, err := s.Database.GetUser(context.Background(), s.Config.CurrentUserName)
		if err != nil {
			return fmt.Errorf("failed to retrieve user from database: %w", err)
		}

		return handler(s, cmd, user)
	}
}

