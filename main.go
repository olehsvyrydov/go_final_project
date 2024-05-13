package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Go application started")
	port := "7540"
	env := os.Getenv("TODO_PORT")
	if env != "" {
		port = env
	}
	storeService := GetStoreService()
	if storeService != nil {
		defer storeService.store.db.Close()
	}

	fmt.Println("Application running with port", port)

	err := ListenApi(port)

	if err != nil {
		fmt.Println(err)
	}

}
