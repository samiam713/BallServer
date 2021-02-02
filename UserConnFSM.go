package main

import (
	"encoding/json"
	"fmt"
	"net"
)

// UserConnState is the state of the UserConnFSM
type UserConnState int

const (
	usUndecided UserConnState = iota
	usScanning
	usCountDown
	usInGame
	usPostGame
)

var clientConnHandlers = [](func(*UserConnFSM, []byte)){undecidedHandler, scanningHandler, countdownHandler, inGameHandler, postGameHandler}

// UserConnFSM is an object representing a connection to a client
type UserConnFSM struct {
	state      UserConnState
	clientConn *net.Conn
	game       *Game
	id         string
}

func (userConnFSM *UserConnFSM) isHost() bool { return userConnFSM.game.hostConn == userConnFSM }

func (userConnFSM *UserConnFSM) opponent() *UserConnFSM {
	if userConnFSM.isHost() {
		return userConnFSM.game.clientConn
	}
	return userConnFSM.game.hostConn
}

func (userConnFSM *UserConnFSM) unlockGame() {
	if userConnFSM.game != nil {
		userConnFSM.game.lock.Unlock()
	}
}

func (userConnFSM *UserConnFSM) lockGame() {
	if userConnFSM.game != nil {
		userConnFSM.game.lock.Lock()

	}
}

func undecidedHandler(userConnFSM *UserConnFSM, message []byte) {
	switch message[0] {
	case cmRequestHosts:
		// write 10 active games back to client
		joinableGames, err := json.Marshal(get10JoinableGames())
		if err != nil {
			panic("JOINABLE GAMES MARSHAL ERROR")
		}
		message := append(sm10Joinable, joinableGames...)
		message = append(message, '\n')
		(*userConnFSM.clientConn).Write(message)

	case cmHostGame:
		userConnFSM.HostGame()
	case cmJoinGame:
		// acquire potential game
		hostID := string(message[2:])
		hostIDtoGameLock.Lock()
		defer hostIDtoGameLock.Unlock()
		game := hostIDToGame[hostID]

		// check if game has been deleted
		if game == nil {
			(*userConnFSM.clientConn).Write(smJoinFailedInactive)
			return
		}

		// check if someone else joined
		if game.clientConn != nil {
			(*userConnFSM.clientConn).Write(smJoinFailedOtherJoined)
			return
		}

		// set up game now that both players in it
		userConnFSM.game = game
		game.lock.Lock()
		game.clientConn = userConnFSM
		game.state = gsBothScanning

		startScanningMessage := append(smStartScanning, []byte(userConnFSM.id)...)
		startScanningMessage = append(startScanningMessage, '\n')

		(*userConnFSM.clientConn).Write(startScanningMessage)
		(*game.hostConn.clientConn).Write(startScanningMessage)

		userConnFSM.state = usScanning
		game.hostConn.state = usScanning
	case cmReset:
		userConnFSM.dellocateGame()

	}
}

func scanningHandler(userConnFSM *UserConnFSM, message []byte) {

	switch message[0] {
	case cmDoneScanning:
		if userConnFSM.game.state == gsOneScanned {
			userConnFSM.game.state = gsCountDown

			userConnFSM.game.clientConn.state = usCountDown
			userConnFSM.game.hostConn.state = usCountDown
			(*userConnFSM.opponent().clientConn).Write(smCountDownStart)
			(*userConnFSM.clientConn).Write(smCountDownStart)
			go userConnFSM.game.startGameLoop()
		} else {
			userConnFSM.game.state = gsOneScanned
			(*userConnFSM.opponent().clientConn).Write(smOpponentScanned)
		}
	case cmReset:
		userConnFSM.dellocateGame()
	}
}

func countdownHandler(userConnFSM *UserConnFSM, message []byte) {

	switch message[0] {
	case cmReset:
		userConnFSM.dellocateGame()
	}

}

func inGameHandler(userConnFSM *UserConnFSM, message []byte) {

	switch message[0] {
	case cmAccelerated:
		var accelerationVector Vector
		err := json.Unmarshal(message[2:], &accelerationVector)
		if err != nil {
			panic("FATAL inGameHandler ERROR")
		}
		
		if userConnFSM.isHost() {
			userConnFSM.game.hostBall.accAccum.adding(accelerationVector)
		} else {
			userConnFSM.game.clientBall.accAccum.adding(accelerationVector)
		}
	case cmReset:
		userConnFSM.dellocateGame()
	}

}

func postGameHandler(userConnFSM *UserConnFSM, message []byte) {

	switch message[0] {
	case cmPlayAgain:
		if userConnFSM.game.state == gsOnePlayAgain {
			userConnFSM.game.state = gsCountDown

			userConnFSM.game.clientConn.state = usCountDown
			userConnFSM.game.hostConn.state = usCountDown
			(*userConnFSM.opponent().clientConn).Write(smCountDownStart)

			go userConnFSM.game.startGameLoop()
		} else {
			userConnFSM.game.state = gsOnePlayAgain
			(*userConnFSM.opponent().clientConn).Write(smOpponentPlayAgain)
		}
	case cmReset:
		userConnFSM.dellocateGame()
	}
}
