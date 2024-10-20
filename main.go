package main

import (
	"fmt"
	"key_store/controllers"
	"key_store/store"
	"log"
	"net/http"
)

const PORT = 8000

func main() {
	store.InitKVStore("data.json")
	mux := http.NewServeMux()

	mux.HandleFunc("/get", controllers.Get)
	mux.HandleFunc("/set", controllers.Set)
	mux.HandleFunc("/update", controllers.Update)
	mux.HandleFunc("/delete", controllers.Delete)

	serverAddr := fmt.Sprintf(":%d", PORT)
	fmt.Printf("Starting server on port: %d\n", PORT)

	err := http.ListenAndServe(serverAddr, mux)
	if err != nil {
		log.Fatalf("Error while starting server: %v", err)
	}
}
