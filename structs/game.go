package structs

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type GamePool struct {
	CreateGame    chan *Game
	DeleteGame    chan *Game
	Games         map[string]*Game
	Broadcast     chan Message
	DirectMessage chan Message
	Mutex         sync.Mutex
}

type Game struct {
	ID      string
	Players map[*Player]*websocket.Conn
	Mutex   sync.Mutex
}

func NewGamePool() *GamePool {
	return &GamePool{
		CreateGame:    make(chan *Game),
		DeleteGame:    make(chan *Game),
		Games:         make(map[string]*Game),
		Broadcast:     make(chan Message),
		DirectMessage: make(chan Message),
	}
}

func (game *Game) evaluateRound() {

	// FIX
	// SHITTY game mechanics. must change these absolute abominations

	// Grab boths moves
	// evaluate the outcome
	// Create individual messages
	// Direct message to players

	if len(game.Players) == 2 {

		var mv1 string
		var mv2 string
		var p []*Player = make([]*Player, 2)
		var cnt int = 0
		for player := range game.Players {
			p[cnt] = player
			cnt++
		}
		mv1 = p[0].Move
		mv2 = p[1].Move

		var playerOneMessage Message = Message{}
		var playerTwoMessage Message = Message{}

		if mv1 == mv2 {
			playerOneMessage.MessageType = "tie"
			playerTwoMessage.MessageType = "tie"
		} else {

			switch {
			case mv1 == "scissors" && mv2 == "rock":
				playerOneMessage.MessageType = "loss"
				playerTwoMessage.MessageType = "win"
			case mv1 == "rock" && mv2 == "scissors":
				playerOneMessage.MessageType = "win"
				playerTwoMessage.MessageType = "loss"
			case mv1 == "paper" && mv2 == "scissors":
				playerOneMessage.MessageType = "loss"
				playerTwoMessage.MessageType = "win"
			case mv1 == "scissors" && mv2 == "paper":
				playerOneMessage.MessageType = "win"
				playerTwoMessage.MessageType = "loss"
			case mv1 == "rock" && mv2 == "paper":
				playerOneMessage.MessageType = "loss"
				playerTwoMessage.MessageType = "win"
			case mv1 == "paper" && mv2 == "rock":
				playerOneMessage.MessageType = "win"
				playerTwoMessage.MessageType = "loss"
			}
		}

		playerOneMessage.MessageBody = Message{MessageType: p[1].Move, MessageBody: p[0]}
		playerTwoMessage.MessageBody = Message{MessageType: p[0].Move, MessageBody: p[1]}

		// Reset Lobby
		p[0].Move = ""
		p[0].Ready = false
		p[1].Move = ""
		p[1].Ready = false

		p[0].GamePool.DirectMessage <- playerOneMessage
		p[1].GamePool.DirectMessage <- playerTwoMessage

	}
}

func (gamePool *GamePool) Run() {
	for {
		select {
		case game := <-gamePool.CreateGame:
			gamePool.Mutex.Lock()
			gamePool.Games[game.ID] = game
			gamePool.Mutex.Unlock()
			// log.Println("New Game has been created, Game ID: ", game.ID)
			// log.Println("Size of Game Pool: ", len(gamePool.Games))
			break
		case game := <-gamePool.DeleteGame:
			gamePool.Mutex.Lock()
			delete(gamePool.Games, game.ID)
			gamePool.Mutex.Unlock()
			// log.Println("Game removed: ", game.ID)
			// log.Println("Size of Game Pool: ", len(gamePool.Games))
			break
		case message := <-gamePool.Broadcast:
			var roomID string = message.MessageBody.(*Game).ID
			for client := range gamePool.Games[roomID].Players {
				if err := client.Conn.WriteJSON(Message{MessageType: message.MessageType, MessageBody: true}); err != nil {
					fmt.Println(err)
					return
				}
			}
		case message := <-gamePool.DirectMessage:
			switch {
			case message.MessageType == "OpponentReady":
				var player *Player = message.MessageBody.(*Player)
				if err := player.Conn.WriteJSON(Message{MessageType: message.MessageType, MessageBody: true}); err != nil {
					fmt.Println(err)
				}
			default:
				var result string = message.MessageType
				var subMsg Message = message.MessageBody.(Message)
				var player *Player = subMsg.MessageBody.(*Player)
				var oppPlayerMove string = subMsg.MessageType

				var replyMessage Message = Message{MessageType: "RoundResult", MessageBody: Message{MessageType: result, MessageBody: Message{MessageType: "OpponentMove", MessageBody: oppPlayerMove}}}
				if err := player.Conn.WriteJSON(replyMessage); err != nil {
					fmt.Println(err)
				}

			}
		}
	}
}
