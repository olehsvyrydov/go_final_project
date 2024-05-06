package main

import (
	"os"
)

func main() {
	port := "7540"
	env := os.Getenv("TODO_PORT")
	if env != "" {
		port = env
	}
	ListenApi(port)

}
