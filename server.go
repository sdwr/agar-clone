package main

import (
    "time"
    "math"
    "math/rand"
    "fmt"
    "os"
    "log"
    "net/http"
    "io/ioutil"
    "encoding/json"
    "strconv"

    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
)

type State struct {
    Size int
    Players map[int]Player
    Pellets []Position
}

type Player struct {
    ID int
    Name string
    Coords Position
    MousePos Position
    Speed float64
    Size int
}

type Position struct {
        x float64
        y float64
}

type Client struct {
    Connection *websocket.Conn
    Name string
    ID int
}

//types are START and MOUSEPOS
type Message struct {
    Sender *Client
    Type string
    mouseX float64
    mouseY float64
    GameState State
}

type Loggerino struct {
    Level LogLevel
}

type LogLevel int
const(
        error LogLevel = 1
        prod LogLevel = 2
        message LogLevel = 3
        micro LogLevel = 4
        debug LogLevel = 5
)

func (l *Loggerino) log(level LogLevel, v ...interface{}) {
       if level <= l.Level {
            log.Println(v)
       }
}

//***************************************************************************
//GLOBAL VARIABLES :)(
//***************************************************************************

var loggerino Loggerino

var lastID int

var gameState State
var lastUpdated time.Time

var clients map[*Client]bool

var incomingMessages []Message

var randomSource *rand.Rand

var upgrader websocket.Upgrader

func initState(state State) {
    state.Size = 100
    addPellets(500, state)
}



func gameLoop(state State) {
    elapsedTime := timeElapsed(lastUpdated)
    minimumLoop, _ := time.ParseDuration("33ms")
    time.Sleep(minimumLoop - elapsedTime)
    elapsedTime = timeElapsed(lastUpdated)
    loggerino.log(micro, "starting game loop after %d ms ",elapsedTime.Milliseconds())
    updatePlayers(state.Players, int(elapsedTime.Milliseconds()))
    checkCollisions()
    lastUpdated = time.Now() 
}

func updatePlayers(players map[int]Player, elapsedMillis int) {
	for _, curr := range players {
	dist := curr.Speed * float64(elapsedMillis/1000)
	dir := addPos(curr.MousePos, negatePos(curr.Coords))
	scaledDir := multPos(normalizeVector(dir),dist)
	curr.Coords = addPos(scaledDir, curr.Coords)
    }
}

func checkCollisions() {
    //needs to be less than n^2
}

func addPellets(amt int, state State) {
    for i:=0; i < amt; i++ {
        state.Pellets = append(state.Pellets, getRandomPos(state.Size, state.Size))
    }
}

func handleIncomingMessages() {
        messages := incomingMessages
        incomingMessages =  []Message{}

        for _, msg := range messages {
            if msg.Type == "START" {
           	gameState.Players[msg.Sender.ID] = Player{
			Name: "",
			Coords: getRandomPos(gameState.Size, gameState.Size),
			MousePos: Position{x:0,y:0,},
			Speed: 10,
			Size: 5,}
		emitId(msg.Sender)
		}
            if msg.Type == "MOUSEPOS" {
		    player := gameState.Players[msg.Sender.ID]
		    player.MousePos = Position{x:msg.mouseX,y:msg.mouseY,}
            }
        }
}

//***************************************************************************
//HELPER FUNCTIONS
//***************************************************************************
func generateID() int {
    lastID++
    return lastID
}

func timeElapsed(prev time.Time) time.Duration {
    currentTime := time.Now()
    return currentTime.Sub(prev)
}

func getRandomPos(maxX int,maxY int) Position {
    var pos Position
    pos.x = randomSource.Float64() * float64(maxX)
    pos.y = randomSource.Float64() * float64(maxY)
    return pos
}

func negatePos(pos Position) Position {
	return Position{x:pos.x*-1, y:pos.y*-1}
}

func addPos(pos1 Position, pos2 Position) Position {
	return Position{pos1.x+pos2.x, pos1.y+pos2.y}
}

func multPos(pos Position, scalar float64) Position {
	return Position{x:pos.x*scalar,y:pos.y*scalar}
}

func normalizeVector(pos Position) Position {
	scalingFactor := 1 / math.Sqrt(pos.x * pos.x + pos.y * pos.y)
	return Position{x:pos.x*scalingFactor, y:pos.y*scalingFactor}
}

//******************************************************************************
//SERVER FUNCTIONS
//******************************************************************************

func indexHandler(w http.ResponseWriter, r *http.Request) {
        body, _:= ioutil.ReadFile("./web/index/html")
        fmt.Fprintf(w, "%s", body)
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        loggerino.log(error, err)
        return
    }
    var client Client
    client.Connection = conn
    client.ID = generateID()
    client.Name = ""
    clients[&client]=true
    listenToSocket(client)
}

func initRouter() *mux.Router {
    router := mux.NewRouter()
    router.HandleFunc("/", indexHandler)
    router.HandleFunc("/socket", socketHandler)
    return router
}


func startServer() {
    router := initRouter()
    loggerino.log(prod, "running server on port 4404")
    log.Fatal(http.ListenAndServe(":4404", router))
}

//***************************************************************************
//SOCKET FUNCTIONS
//***************************************************************************
func emitId(client *Client) {
	message := Message{Sender:client,Type:"ID",}
	encodedMessage, err := json.Marshal(message)
	if(err != nil) {
        loggerino.log(error, err)
		return
	}
	loggerino.log(prod, "gave client %v ID %d", &client, client.ID)
	client.Connection.WriteMessage(websocket.TextMessage,encodedMessage)
}

func broadcastState(state State) {
	message := Message{Type:"STATE",GameState:state,}
	encodedMessage, err := json.Marshal(message)
	if(err != nil) {
            loggerino.log(error, err)
            return
    }
    loggerino.log(micro, "broadcasting %v to each client:", state)
    for client, _ := range clients {
        client.Connection.WriteMessage(websocket.TextMessage, encodedMessage)
	loggerino.log(micro, "broadcast to client %v", client)
    }
}
//does mux router make a subroutine for this?
//gonna have to test this out
func listenToSocket(client Client) {
    for {
          _, msg, err := client.Connection.ReadMessage()
          if err != nil {
                  break
          }
         var message Message
         json.Unmarshal(msg, &message)
	     message.Sender = &client
         incomingMessages = append(incomingMessages, message)
         loggerino.log(micro, "added message to incoming queue")
         loggerino.log(micro, message)

   }
}

//***************************************************************************
//MAIN FUNCTIONS
//***************************************************************************

func initGlobals() {
    loggerino = Loggerino{Level:prod}
    clients = make(map[*Client]bool)
    gameState.Players = make(map[int]Player)
    randomSource = rand.New(rand.NewSource(99))
    lastID = 0
    upgrader = websocket.Upgrader{}
    upgrader.CheckOrigin = func(r *http.Request) bool {return true}
}

func runWrapper() {
    for {
	run()
    }
}

func run() {
    gameLoop(gameState)
    broadcastState(gameState)
    handleIncomingMessages()
}


func main() {
    initGlobals()
    args := os.Args[1:]
    level, _ := strconv.Atoi(args[0])
    loggerino.Level = LogLevel(level)
    initState(gameState)
    lastUpdated = time.Now()
    loggerino.log(prod, "Starting game loop")
    go runWrapper()
    startServer()
}
