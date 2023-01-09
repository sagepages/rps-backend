package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	id "github.com/sagepages/rps-backend/id"
	structs "github.com/sagepages/rps-backend/structs"
)

// FIX - Update origin after completion

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebsocketHandler(pool *structs.Pool, gamePool *structs.GamePool, w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	var player *structs.Player = &structs.Player{ID: id.RandStr(20), Ready: false, Conn: conn, Pool: pool, GamePool: gamePool}

	pool.Register <- player
	player.Read()
}
