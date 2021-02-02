package main

import (
	"encoding/json"
	"math"
	"strconv"
	"sync"
	"time"
)

const (
	gameInitialRadius   = 0.5
	gameInitialDiameter = gameInitialRadius * 2
	ballInitialY        = gameInitialRadius * 0.75
)

// GameState is state of game
type GameState int

const (
	gsBothScanning GameState = iota
	gsOneScanned
	gsCountDown
	gsPlayingGame
	gsPostGame
	gsOnePlayAgain
)

const (
	gameLength float64 = 45.0
)

var hostIDtoGameLock sync.Mutex
var hostIDToGame = map[string]*Game{}

func get10JoinableGames() []string {
	joinableGames := make([]string, 0)

	var count = 0
	hostIDtoGameLock.Lock()
	for id, game := range hostIDToGame {
		if game.clientConn == nil {
			joinableGames = append(joinableGames, id)
			count++
			if count == 10 {
				break
			}
		}
	}
	hostIDtoGameLock.Unlock()

	return joinableGames
}

// Game represents all data necessary for a ball game
type Game struct {
	hostBall, clientBall Ball
	every60th            *time.Ticker
	cancelEvery60th      chan interface{}
	hostConn, clientConn *UserConnFSM
	startNanosecond      int64
	endNanosecond        int64
	currentNanosecond    int64
	boundaryRadius       float64
	state                GameState
	lock                 sync.Mutex
}

// HostGame creates and registers a game
func (user *UserConnFSM) HostGame() {
	user.game = &Game{hostBall: hostBall, clientBall: clientBall, cancelEvery60th: make(chan interface{}), hostConn: user, boundaryRadius: 0.5}
	user.game.lock.Lock()
	hostIDtoGameLock.Lock()
	hostIDToGame[user.id] = user.game
	hostIDtoGameLock.Unlock()
}

func (user *UserConnFSM) dellocateGame() {

	if user.game == nil {
		return
	}

	// remove all references
	hostIDtoGameLock.Lock()
	delete(hostIDToGame, user.game.hostConn.id)
	hostIDtoGameLock.Unlock()

	if user.game.every60th != nil {
		user.game.cancelEvery60th <- nil
	}

	if opponent := user.opponent(); opponent != nil {
		(*opponent.clientConn).Write(smOpponentLeft)
		opponent.game = nil
		opponent.state = usUndecided
	}

	user.game.lock.Unlock()

	user.game = nil
	user.state = usUndecided

}

func (g *Game) nanoSecondsSinceStart() int64 {
	return g.currentNanosecond - g.startNanosecond
}

func (g *Game) updateBoundaryRadius() {
	gameProportion := float64(g.currentNanosecond-g.startNanosecond) / (gameLength * 1000000000)
	g.boundaryRadius = gameInitialRadius * math.Sqrt(1.0-gameProportion)
}

func (g *Game) checkBoundaryCollision(ball *Ball) bool {
	return g.boundaryRadius < (ball.pos.magnitude() + ballRadius)
}

func (g *Game) startGameLoop() {

	// reset game balls
	g.hostBall = hostBall
	g.clientBall = clientBall

	// set up ticker
	g.every60th = time.NewTicker(16666667)
	ticker := g.every60th.C

	// update times
	g.currentNanosecond = time.Now().UnixNano()
	g.startNanosecond = g.currentNanosecond + 5*1000000000
	g.endNanosecond = g.startNanosecond + 45*1000000000

	for {
		var time time.Time
		var ok bool
		select {
		case time, ok = <-ticker:
		case <-g.cancelEvery60th:
			ok = false
		}

		g.lock.Lock()

		// check to see if game is over
		if !ok {
			// the game is over so return
			break
		}

		// InGame || PregameCountdown
		dt := float64(time.UnixNano()-g.currentNanosecond) / 1000000000
		g.currentNanosecond = time.UnixNano()

		// check if we're still in pregame
		if g.state == gsCountDown {
			nanoSecondsSinceStart := g.nanoSecondsSinceStart()
			if nanoSecondsSinceStart < 0 {
				// keep publishing pregameMessages
				pregameTimeMessage := append(smCountDownTime, strconv.Itoa(-int(nanoSecondsSinceStart))...)
				pregameTimeMessage = append(pregameTimeMessage, '\n')
				(*g.clientConn.clientConn).Write(pregameTimeMessage)
				(*g.hostConn.clientConn).Write(pregameTimeMessage)
			} else {
				// start publishing playingMessages
				(*g.clientConn.clientConn).Write(smStartGame)
				(*g.hostConn.clientConn).Write(smStartGame)
				g.clientConn.state = usInGame
				g.hostConn.state = usInGame
				g.state = gsPlayingGame
			}
			g.lock.Unlock()
			continue
		}

		// check if game is over (should never execute as collision will happen first)
		if g.currentNanosecond > g.endNanosecond {
			panic("How'd we get this far without collision?")
			// you could default to client winning?
		}

		gameOver := g.updateGameStateAndPublish(dt)

		if gameOver {
			break
		}
		g.lock.Unlock()
	}

	g.state = gsPostGame
	g.every60th.Stop()
	g.every60th = nil

	g.lock.Unlock()
}

// returns true iff gameOver
func (g *Game) updateGameStateAndPublish(dt float64) bool {
	if g.state != gsPlayingGame {
		panic("Invalid game state for the ticker")
	}

	// else advance game logic:

	// BALLS
	g.hostBall.advanceBall(dt)
	g.clientBall.advanceBall(dt)
	collideIfNecessary(&g.hostBall, &g.clientBall)

	// BOUNDARY
	g.updateBoundaryRadius()

	if g.checkBoundaryCollision(&g.hostBall) {
		(*g.clientConn.clientConn).Write(smClientWon)
		(*g.clientConn).state = usPostGame
		(*g.hostConn.clientConn).Write(smClientWon)
		(*g.hostConn).state = usPostGame
		g.state = gsPostGame
		return true
	}
	if g.checkBoundaryCollision(&g.clientBall) {
		(*g.clientConn.clientConn).Write(smHostWon)
		(*g.clientConn).state = usPostGame
		(*g.hostConn.clientConn).Write(smHostWon)
		(*g.hostConn).state = usPostGame
		g.state = gsPostGame
		return true
	}

	// write current position to client and host
	gameStateJSON, err := json.Marshal(smGameStateJSON{Host: g.hostBall.pos, Client: g.clientBall.pos, Radius: g.boundaryRadius})
	if err != nil {
		panic("COULDN'T ENCODE JSON")
	}

	message := append(smGameState, gameStateJSON...)
	message = append(message, '\n')

	(*g.clientConn.clientConn).Write(message)
	(*g.hostConn.clientConn).Write(message)
	return false
}
