package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/GravitiMusic/gator/internal/config"
	"github.com/GravitiMusic/gator/internal/database"

	_ "github.com/lib/pq"
)

func main() {
	c, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}

	db, err := sql.Open("postgres", c.DbURL)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}
	defer db.Close()

	dbQueries := database.New(db)

	ste := &state{Config: &c, Database: dbQueries}
	cmds := &commands{CommandMap: make(map[string]func(*state, command) error)}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", handlerAddfeed)
	cmds.register("feeds", handlerFeeds)

	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Println("Usage: gator <command> [args...]")
		os.Exit(1)
	}

	cmd := command{Name: args[0], Args: args[1:]}
	err = cmds.run(ste, cmd)
	if err != nil {
		fmt.Println("Error executing command:", err)
	}
}