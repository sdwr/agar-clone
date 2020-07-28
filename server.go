package main

import (
    "time"
    "math"
    "math/rand"
    "fmt"
    "log"
    "net/http"
    "io/ioutil"
    "encoding/json"

    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
)

type State struct {
    Size int
    Players map[*Client]Player
    Pellets []Position
}

type Player struct {
    ID int
    Name string
    Coords Position
    MousePos Position
    Speed float32
    Size int
}

type Position struct {
        x float32
        y float32
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
    mouseX float32
    mouseY float32
    GameState State
}

//***************************************************************************
//GLOBAL VARIABLES :)(
//***************************************************************************

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
    currentTime := time.Now()
    elapsedTime := currentTime.Sub(lastUpdated).Milliseconds()
    updatePlayers(state.Players, int(elapsedTime))
    checkCollisions()
    lastUpdated = currentTime
}

func updatePlayers(players map[*Client]Player, elapsedMillis int) {
	for _, curr := range players {
	dist := curr.Speed * float32(elapsedMillis/1000)
	dir := addPos(curr.MousePos, negatePos(curr.Coords))
	angle := math.Atan(float64(dir.y/dir.x))
	if dir.x < 0 {
		angle += math.Pi
	}
	if angle > 2*math.Pi {
		angle -= 2*math.Pi
	}
	if angle < 0 {
		angle += 2*math.Pi
	}
	scaledDir := Position{x:dist*float32(math.Cos(angle)),y:dist*float32(math.Sin(angle))}
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
           	gameState.Players[msg.Sender] = Player{
			Name: "",
			Coords: getRandomPos(gameState.Size, gameState.Size),
			MousePos: Position{x:0,y:0,},
			Speed: 10,
			Size: 5,}
		emitId(msg.Sender)
		}
            if msg.Type == "MOUSEPOS" {
		    player := gameState.Players[msg.Sender]
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

func getRandomPos(maxX int,maxY int) Position {
    var pos Position
    pos.x = randomSource.Float32() * float32(maxX)
    pos.y = randomSource.Float32() * float32(maxY)
    return pos
}

func negatePos(pos Position) Position {
	return Position{x:pos.x*-1, y:pos.y*-1}
}

func addPos(pos1 Position, pos2 Position) Position {
	return Position{pos1.x+pos2.x, pos1.y+pos2.y}
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
        log.Println(err)
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
    log.Println("running server")
    log.Fatal(http.ListenAndServe(":4404", router))
}

//***************************************************************************
//SOCKET FUNCTIONS
//***************************************************************************
func emitId(client *Client) {
	message := Message{Sender:client,Type:"ID",}
	encodedMessage := json.Marshal(message)
	if(err != nil) {
		log.Println(err)
		return
	}
	client.Connection.WriteMessage(websocket.TextMessage,encodedMessage)
}

func broadcastState(state State) {
	message := Message{Type:"STATE",GameState:state,}
	encodedMessage := json.Marshal(message)
	if(err != nil) {
	    log.Println(err)
            return
    }
    for client, _ := range clients {
        client.Connection.WriteMessage(websocket.TextMessage, encodedMessage)
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

   }
}

//***************************************************************************
//MAIN FUNCTIONS
//***************************************************************************

func initGlobals() {
    clients = make(map[*Client]bool)
    gameState.Players = make(map[*Client]Player)
    randomSource = rand.New(rand.NewSource(99))
    lastID = 0
    upgrader = websocket.Upgrader{}
}

func run() {
    gameLoop(gameState)
    broadcastState(gameState)
    handleIncomingMessages()
}


func main() {
    initGlobals()
    startServer()
    initState(gameState)
    lastUpdated = time.Now()
    for {
            run()
    }
}
