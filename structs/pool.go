package structs

import (
	"fmt"
)

type Pool struct {
	Register   chan *Player
	Unregister chan *Player
	Players    map[*Player]bool
	Broadcast  chan Message
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Player),
		Unregister: make(chan *Player),
		Players:    make(map[*Player]bool),
		Broadcast:  make(chan Message),
	}
}

func (pool *Pool) Run() {

	for {
		select {
		case player := <-pool.Register:
			pool.Players[player] = true
			// log.Println("New Connection, ID: ", player.ID)
			// log.Println("Connection pool size: ", len(pool.Players))
			break
		case player := <-pool.Unregister:
			delete(pool.Players, player)
			// log.Println("Connection pool size: ", len(pool.Players))
			break
		case message := <-pool.Broadcast:
			for client := range pool.Players {
				if err := client.Conn.WriteJSON(message); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}
