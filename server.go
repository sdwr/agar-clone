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
    "strconv"

    "github.com/gorilla/mux"

    "github.com/sdwr/agar-clone/socket"
    "github.com/sdwr/agar-clone/auth"
    . "github.com/sdwr/agar-clone/types"
)

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

var room *socket.Room

var gameState State
var lastUpdated time.Time

var randomSource *rand.Rand

var router *mux.Router

func initState() {
    gameState.Size = 1000
    gameState.PlayerSize = 20
    gameState.ChunkSize = 250
    createChunks()
    addPellets(500)
}

//chunksize must be no greater than 1/100th of the map size 
func createChunks() {
    gameState.Chunks = make(map[int][]Object)
    for x:= 0; x < int(gameState.Size); x+= gameState.ChunkSize {
	for y:=0; y < int(gameState.Size); y += gameState.ChunkSize {
		gameState.Chunks[posToChunkIndex(Position{X:float64(x),Y:float64(y)})] = make([]Object,1,10)
	}
    }
}

func posToChunkIndex(pos Position) int {
    return (int(pos.X)/gameState.ChunkSize) * 100 + (int(pos.Y)/gameState.ChunkSize)
}

func addToChunk(o Object) {
    chunk := gameState.Chunks[posToChunkIndex(o.Coords)]
    gameState.Chunks[posToChunkIndex(o.Coords)] = append(chunk, o)
}

func removeFromChunk(o Object) {
	chunkIndex := posToChunkIndex(o.Coords)
	chunk := gameState.Chunks[chunkIndex]
	for i, v := range chunk {
	    if v == o {
		chunk[i] = chunk[len(chunk)-1]
	        gameState.Chunks[chunkIndex] = chunk[:len(chunk)-1]
	    }
	}
}

func getSurroundingChunks(pos Position) []Object {
	objects := make([]Object,0,100)
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			chunkPos := addPos(pos, Position{X:float64(x*gameState.ChunkSize),Y:float64(y*gameState.ChunkSize)})
			chunk := gameState.Chunks[posToChunkIndex(chunkPos)]
			objects = append(objects, chunk...)
		}
	}
	return objects
}

func createPlayer(id int) Player {
	return Player{	Type: "PLAYER",
			ID: id,
			Name: "",
			Coords: getRandomPos(int(gameState.Size), int(gameState.Size)),
			MousePos: Position{X:0,Y:0,},
			Size: gameState.PlayerSize,
			Speed: 0,}

}

func createPlayerObject(p Player) Object {
	return Object{ Type: "PLAYER",
			ID: p.ID,
			Size: p.Size,
			Coords: p.Coords,}
}

func createPlayerObjects() []Object {
	pos := make([]Object,0,10)
	for _, p := range gameState.Players {
		pos = append(pos, createPlayerObject(p))
	}
	return pos
}

func removePlayer(id int) {
    delete(gameState.Players, id)
}

func gameLoop() {
    elapsedTime := timeElapsed(lastUpdated)
    if elapsedTime.Milliseconds() > 30 {
	    loggerino.log(message, "Warning: last frame took %d ms to run", elapsedTime.Milliseconds())
    }
    loggerino.log(micro, elapsedTime.Milliseconds())
    minimumLoop, _ := time.ParseDuration("33ms")
    time.Sleep(minimumLoop - elapsedTime)
    elapsedTime = timeElapsed(lastUpdated)
    loggerino.log(micro, "starting game loop after %d ms ",elapsedTime.Milliseconds())
    updatePlayers(int(elapsedTime.Milliseconds()))
    checkCollisions()
    lastUpdated = time.Now()
}

func updatePlayers(elapsedMillis int) {
	players := gameState.Players
    for key, curr := range players {
	    if curr.RespawnMillis > 0 {
		calculateRespawn(curr, elapsedMillis)
		curr = gameState.Players[curr.ID]
	    }
            dist := curr.Speed * float64(elapsedMillis)
	    dir := addPos(curr.MousePos, negatePos(curr.Coords))
	    scaledDir := multPos(normalizeVector(dir),dist)
	    curr.Coords = addPos(scaledDir, curr.Coords)
	    //check wall boundary
	    gameSizeFloat := float64(gameState.Size)
	    halfSize := float64(curr.Size/2)
	    if curr.Coords.X < 0+halfSize {
		curr.Coords.X = 0+halfSize
	    }
	    if curr.Coords.X > gameSizeFloat-halfSize {
		curr.Coords.X = gameSizeFloat-halfSize
	    }
	    if curr.Coords.Y < 0+halfSize {
		curr.Coords.Y = 0+halfSize
	    }
	    if curr.Coords.Y > gameSizeFloat-halfSize {
		curr.Coords.Y = gameSizeFloat-halfSize
	    }
	    gameState.Players[key]=curr
    }
}

func calculateRespawn(p Player, millis int) {
	p.RespawnMillis -= float64(millis)
	gameState.Players[p.ID]=p
	if p.RespawnMillis <= 0 {
		p.Size = gameState.PlayerSize
		p.Coords = getRandomPos(int(gameState.Size), int(gameState.Size))
		gameState.Players[p.ID]=p
		calculateSpeed(p)
	}
}

func checkCollision(p Player, o Object) bool {
	pLeft := p.Coords.X-p.Size/2
	pRight := p.Coords.X+p.Size/2
	pUp := p.Coords.Y-p.Size/2
	pDown := p.Coords.Y + p.Size/2

	oLeft := o.Coords.X-o.Size/2
	oRight := o.Coords.X+o.Size/2
	oUp := o.Coords.Y-o.Size/2
	oDown := o.Coords.Y + o.Size/2

	xOverlap := 0
	yOverlap := 0
	if oLeft < pRight && oLeft > pLeft {
		xOverlap = 1
	}else if pLeft < oRight && pLeft > oLeft {
		xOverlap = 1
	}
	if xOverlap > 0 {
		if pUp < oDown && pUp > oUp {
			yOverlap = 1
		} else if oUp < pDown && oUp > pUp {
			yOverlap = 1
		}
		if xOverlap > 0 && yOverlap > 0 {
			return true
		}
	}
	return false
}

func checkCollisions() {
	players := gameState.Players
	for _, curr := range players {
	//check obj collisions
	chunkIndex := posToChunkIndex(curr.Coords)
	chunk := gameState.Chunks[chunkIndex]
		for _, o := range chunk {
		   if checkCollision(curr, o) && curr.ID != o.ID {
			resolveCollision(curr, o)
		   }
		}
	//check player collisions n^2 :(
		for _, other := range players {
			o := createPlayerObject(other)
			if curr.ID != o.ID && checkCollision(curr, o) {
				resolveCollision(curr, o)
			}
		}
	}
}

func resolveCollision(p Player, o Object) {
	if(o.Type == "PELLET") {
		removeFromChunk(o)
		increaseSize(p)
		createPellet()
	}
	if(o.Type == "PLAYER") {
		if o.Size > p.Size {
			killPlayer(p)
		} else if p.Size > o.Size {
			killPlayer(gameState.Players[o.ID])
		}
	}
}

func killPlayer(p Player) {
	p.Coords = Position{gameState.Size/2,float64(-20)}
	p.Size = 0
	p.Speed = 0
	p.RespawnMillis = 5000
	gameState.Players[p.ID] = p
}

func addPellets(amt int) {
    for i:=0; i < amt; i++ {
	    createPellet()
    }
}

func createPellet() {
	    pellet := Object{Type:"PELLET", Size: 5, ID: -1, Coords: getRandomPos(int(gameState.Size), int(gameState.Size))}
	    addToChunk(pellet)

}

func increaseSize(p Player) {
	p.Size++
	gameState.Players[p.ID]=p
	calculateSpeed(p)
}

func calculateSpeed(p Player) {
	p.Speed = (1 / float64(p.Size)) + 0.06
	gameState.Players[p.ID]=p
}

func handleIncomingMessage(msg *socket.Message) {
    if msg.Type == "START" {
	player := createPlayer(msg.Sender.ID)
	gameState.Players[msg.Sender.ID] = player
	calculateSpeed(player)
	emitID(msg.Sender)
    } else if msg.Type == "MOUSEPOS" {
	player := gameState.Players[msg.ID]
	player.MousePos = Position{X:msg.Payload.MouseX, Y:msg.Payload.MouseY,}
	gameState.Players[msg.ID] = player
    } else if msg.Type == "CREATEBOT" {
	loggerino.log(prod, "bot created")
	createBot()
    }
}

//***************************************************************************
//TESTING FUNCTIONS
//***************************************************************************
func createBot(){
    bot := createPlayer(generateID())
    gameState.Players[bot.ID] = bot
    calculateSpeed(bot)
    go moveBot(bot.ID)
}

func moveBot(id int) {
        loopTime, _ := time.ParseDuration("2s")
    for {
	botMouse := getRandomPos(int(gameState.Size), int(gameState.Size))
	payload := Payload{MouseX: botMouse.X, MouseY: botMouse.Y}
	botMove := socket.Message{Type:"MOUSEPOS", ID:id, Payload:payload,}
	room.FakeIncomingMessage(&botMove)
        time.Sleep(loopTime)
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
    pos.X = randomSource.Float64() * float64(maxX)
    pos.Y = randomSource.Float64() * float64(maxY)
    return pos
}

func negatePos(pos Position) Position {
	return Position{X:pos.X*-1, Y:pos.Y*-1}
}

func addPos(pos1 Position, pos2 Position) Position {
	return Position{pos1.X+pos2.X, pos1.Y+pos2.Y}
}

func multPos(pos Position, scalar float64) Position {
	return Position{X:pos.X*scalar,Y:pos.Y*scalar}
}

func normalizeVector(pos Position) Position {
	scalingFactor := 1 / math.Sqrt(pos.X * pos.X + pos.Y * pos.Y)
	return Position{X:pos.X*scalingFactor, Y:pos.Y*scalingFactor}
}

//******************************************************************************
//SERVER FUNCTIONS
//******************************************************************************

func indexHandler(w http.ResponseWriter, r *http.Request) {
        body, _:= ioutil.ReadFile("./web/index.html")
        fmt.Fprintf(w, "%s", body)
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
    socket.ServeWs(room, w, r)
}

func initRouter() {
    router = mux.NewRouter()
   }

func addRoutes() {
    router.HandleFunc("/socket", socketHandler)
    router.PathPrefix("/web/").Handler(http.StripPrefix("/web/", http.FileServer(http.Dir("./web/"))))
    router.HandleFunc("/", indexHandler)
}

func startServer() {
    loggerino.log(prod, "running server on port 4404")
    log.Fatal(http.ListenAndServe(":4404", router))
}

//***************************************************************************
//SOCKET FUNCTIONS
//***************************************************************************
func removeClient(client *socket.Client) {
    removePlayer(client.ID)
}
func emitID(client *socket.Client) {
    message := socket.Message{ID:client.ID,Type:"ID",Reciever:client}
    room.BroadcastMessage(&message)
}

func broadcastState() {
	outgoingState := OutgoingState{Size:gameState.Size}
	payload := Payload{State:outgoingState,}
	message := socket.Message{Type:"STATE",Payload:payload,}
    loggerino.log(micro, "broadcasting %v to each client:", gameState)
    for client, _ := range room.Clients {
	message.Payload.State.CurrentPlayer = gameState.Players[client.ID]
	message.Payload.State.Objects = getSurroundingChunks(message.Payload.State.CurrentPlayer.Coords)
	message.Payload.State.Objects = append(message.Payload.State.Objects, createPlayerObjects()...)
	message.Reciever = client
	room.BroadcastMessage(&message)
	loggerino.log(micro, "broadcast to client %v", client)
    }
}

//***************************************************************************
//MAIN FUNCTIONS
//***************************************************************************

func initGlobals() {
    loggerino = Loggerino{Level:prod}
    gameState.Players = make(map[int]Player)
    randomSource = rand.New(rand.NewSource(99))
    lastID = 0
}

//crashing with 10+ players because these callbacks are async
//and overlapping the player reads/writes
//TODO: make a queue for these bad boys (mousePos updates)
func initRoom() {
    room = socket.NewRoom()
    room.SetIncomingCallback(handleIncomingMessage)
    room.SetRegisterCallback(emitID)
    room.SetUnregisterCallback(removeClient)
    go room.Run()
}

func setLogLevel() {
    args := os.Args[1:]
    level, _ := strconv.Atoi(args[0])
    loggerino.Level = LogLevel(level)

}

func runWrapper() {
    for {
	run()
    }
}

func run() {
    gameLoop()
    broadcastState()
}

func main() {
    initGlobals()
    initRoom()
    setLogLevel()
    initState()
    lastUpdated = time.Now()
    loggerino.log(prod, "Starting game loop")
    go runWrapper()
    initRouter()
    auth.LoadAuth(router)
    addRoutes()
    startServer()
}
