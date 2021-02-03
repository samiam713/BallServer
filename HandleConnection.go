package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

var activeConnectionsLock sync.Mutex
var activeConnections = make(map[string]bool)

func handleConnection(c *net.Conn) {

	fmt.Println("New Connection")

	scanner := bufio.NewScanner(*c)

	var id string

	// loop until valid username
	for {
		if scanner.Scan() {
			id = scanner.Text()
		} else {
			return
		}

		activeConnectionsLock.Lock()
		// if we already have this id, tell the user
		if activeConnections[id] {
			(*c).Write(smDuplicateUser)
			activeConnectionsLock.Unlock()
			// else tell establish the user is logged in
		} else {
			(*c).Write(smGoodUsername)

			// THIS IS PAIRED WITH DELETION FROM SET
			activeConnections[id] = true
			activeConnectionsLock.Unlock()
			break
		}
	}

	// set up clientConnection using incomingMessages
	userConnection := UserConnFSM{state: usUndecided, clientConn: c, game: nil, id: id}

	for scanner.Scan() {
		userConnection.lockGame()
		if len(scanner.Bytes()) != 0 {
			(clientConnHandlers[userConnection.state])(&userConnection, scanner.Bytes())
		}
		userConnection.unlockGame()
	}

	// the user disconnected, ending game
	userConnection.lockGame()
	userConnection.dellocateGame()

	activeConnectionsLock.Lock()
	delete(activeConnections, id)
	activeConnectionsLock.Unlock()
}
