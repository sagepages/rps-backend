package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	structs "github.com/sagepages/rps-backend/structs"
	ws "github.com/sagepages/rps-backend/ws"
)

func setupRoutes() {

	// Initialize empty pool of Clients
	var pool *structs.Pool = structs.NewPool()

	// Initialize empty pool of Games
	var gamePool *structs.GamePool = structs.NewGamePool()

	go pool.Run()
	go gamePool.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.WebsocketHandler(pool, gamePool, w, r)
	})
}

func main() {

	// Seeding
	rand.Seed(time.Now().UnixNano())
	fmt.Println("Server started at port: 8080")

	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
