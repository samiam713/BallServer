package main

// A MESSAGE IS A SHORT SEQUENCE OF BYTES SENT OVER THE TCP CONNECTION TO COMMUNICATE TO THE HOST/CLIENT THAT
// AN EVENT (A CHANGE IN THE STATE OF THE PROGRAM) HAS OCCURED ON THE CLIENT/HOST

// client to server messages

// accompanied by vector
const cmAccelerated byte = 'a'

const cmHostGame byte = 'b'

// accompanied by hostID
const cmJoinGame byte = 'c'

const cmDoneScanning byte = 'd'

const cmPlayAgain byte = 'e'

const cmRequestHosts byte = 'f'

const cmReset byte = 'g'

// server to client messages

// sends smGameStateJSON
var smGameState = []byte{'a', ' '}

type smGameStateJSON struct {
	Host, Client Vector
	Radius       float64
}

// only have seen below used when opponent leaves
// var smReset = []byte{'b', '\n'}

// sends a username
var smStartScanning = []byte{'b', ' '}

var smJoinFailedInactive = []byte{'c', '\n'}

var smJoinFailedOtherJoined = []byte{'d', '\n'}

var smDuplicateUser = []byte{'e', '\n'}

var smOpponentLeft = []byte{'f', '\n'}

// sends the time in nanoseconds
var smCountDownTime = []byte{'g', ' '}

var smStartGame = []byte{'h', '\n'}

var smClientWon = []byte{'i', '\n'}
var smHostWon = []byte{'j', '\n'}

var smCountDownStart = []byte{'k', '\n'}

var smGoodUsername = []byte{'l', '\n'}

var smOpponentScanned = []byte{'m', '\n'}

var smOpponentPlayAgain = []byte{'n', '\n'}

// sends []string
var sm10Joinable = []byte{'o', ' '}
