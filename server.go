package main

import (
    "time"
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
    Players []Player
    Pellets []Position
}

type Player struct {
    ID int 
    Name string
    Coords []Position
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
    //addPellets(state)
    lastUpdated = currentTime
}

func updatePlayers(players []Player, elapsedMillis int) {
    for i:=0; i < len(players); i++ {
        // update coords here (trig :()
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
            }
            if msg.Type == "MOUSEPOS" {

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

func broadcastState(state State) {
    encodedState, err := json.Marshal(state)
    if(err != nil) {
            log.Println(err)
            return
    }
    for client, _ := range clients {
        client.Connection.WriteMessage(websocket.TextMessage, encodedState)
    }
}

func listenToSocket(client Client) {
    for {
            _, msg, err := client.Connection.ReadMessage()
            if err != nil {
                    break
            }
           var message Message
           json.Unmarshal(msg, &message)
           incomingMessages = append(incomingMessages, message)

   }
}

//***************************************************************************
//MAIN FUNCTIONS
//***************************************************************************

func initGlobals() {
    clients = make(map[*Client]bool)
    randomSource = rand.New(rand.NewSource(99))
    lastID = 0
    upgrader = websocket.Upgrader{}
}

func run(state State) {
    gameLoop(state)
    broadcastState(state)
    handleIncomingMessages()
}


func main() {
    initGlobals()
    startServer()
    initState(gameState)
    lastUpdated = time.Now()
    for {
            run(gameState)
    }
}
