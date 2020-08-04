package types

type State struct {
    Size float64
    ChunkSize int
    PlayerSize float64
    Chunks map[int][]Object
    Players map[int]Player
}

type OutgoingState struct {
    Size float64
    CurrentPlayer Player
    Objects []Object
}

type Object struct {
    Type string
    Size float64
    ID int
    Coords Position
}

type Player struct {
    Type string
    ID int
    Name string
    Coords Position
    MousePos Position
    Speed float64
    Size float64
    RespawnMillis float64
}

type Position struct {
    X float64
    Y float64
}

type Payload struct {
    MouseX float64
    MouseY float64
    State OutgoingState
}
