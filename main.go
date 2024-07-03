package main

import (
	// "fmt"
	"log"

	"github.com/harshgupta9473/goBank/component"
)

func main() {
	store, err := component.NewPostgressStore()
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("%+v\n", store)
	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	server := component.NewAPIServer(":3000", store)
	server.Run()
}
