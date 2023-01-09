package structs

import (
	"log"

	"github.com/gorilla/websocket"
)

type Player struct {
	ID       string
	Conn     *websocket.Conn
	Pool     *Pool
	GamePool *GamePool
	Ready    bool
	Move     string
}

func (currentPlayer *Player) Read() {

	defer func() {
		currentPlayer.Pool.Unregister <- currentPlayer
		currentPlayer.Conn.Close()
	}()
	for {

		var msg Message = Message{}
		err := currentPlayer.Conn.ReadJSON(&msg)

		if err != nil {
			log.Println(err)
			return
		}

		switch {
		case msg.MessageType == "AddToRoom": // Expects: Message{MessageType: "AddToRoom", "MessageBody": "roomID"}

			// FIX

			// 01/05/2023 - Possibility that a user enters any roomID as a URL via the frontend.
			// Should consider a different approach or handle those scenarios.

			// Check if room already exists. If it does, then there must already have been
			// a currentPlayer that created the room and is inside.
			// Otherwise, create a new room.

			var roomID string = msg.MessageBody.(string)
			if game, exists := currentPlayer.GamePool.Games[roomID]; exists {

				game.Mutex.Lock()
				game.Players[currentPlayer] = currentPlayer.Conn
				game.Mutex.Unlock()

				// Broadcast "RoomIsReady" command
				message := Message{MessageType: "RoomIsReady", MessageBody: currentPlayer.GamePool.Games[roomID]}
				currentPlayer.GamePool.Broadcast <- message
			} else {

				var newGame *Game = &Game{ID: roomID}
				newGame.Players = make(map[*Player]*websocket.Conn)
				newGame.Mutex.Lock()
				newGame.Players[currentPlayer] = currentPlayer.Conn
				newGame.Mutex.Unlock()

				// Send newly created Game
				currentPlayer.GamePool.CreateGame <- newGame
			}

		case msg.MessageType == "Move": // Expects : {MessageType: "Move", MessageBody: {"MessageType": "gameID", MessageBody: "scissor"}}

			var msgMap map[string]interface{} = msg.MessageBody.(map[string]interface{})
			var gameID string = msgMap["type"].(string)
			var players []*Player = make([]*Player, 2)
			var cnt int = 0
			for p := range currentPlayer.GamePool.Games[gameID].Players {
				players[cnt] = p
				cnt++
			}
			currentPlayer.Ready = true
			currentPlayer.Move = msgMap["body"].(string)

			switch {
			case players[0].Ready && players[1].Ready:
				currentPlayer.GamePool.Games[gameID].evaluateRound()
			case !players[0].Ready:
				var readyMessage Message = Message{MessageType: "OpponentReady", MessageBody: players[0]}
				currentPlayer.GamePool.DirectMessage <- readyMessage
			case !players[1].Ready:
				var readyMessage Message = Message{MessageType: "OpponentReady", MessageBody: players[1]}
				currentPlayer.GamePool.DirectMessage <- readyMessage
			}

		case msg.MessageType == "GameOver":
			var roomID string = msg.MessageBody.(string)
			currentPlayer.GamePool.DeleteGame <- currentPlayer.GamePool.Games[roomID]
		}

	}
}
