console.log("file loaded")
//GLOBAL CONSTANTS
const SERVER_URL = "localhost:4404"

const canvas = document.getElementById("game-canvas");
const playerInfo = document.getElementById("player-info");
const mouseInfo = document.getElementById("mouse-info");
const messageInfo = document.getElementById("message-info");
const ctx = canvas.getContext('2d');
const socket = new WebSocket("ws://"+SERVER_URL+"/socket");

const windowSize = 800;


let gameState = {Players:[]};
let id = 1;

let playerCoords = {X:0,Y:0}
let mousePos = {X:0,Y:0}
//DEFINITIONS
class Position {
	constructor(x, y) {
		this.X = x;
		this.Y = y;
	}

}

//CANVAS
canvas.addEventListener('mousemove', function(e) {
	mousePos = getMousePos(e)

}, false);

function getMousePos(e) {
	let rect = canvas.getBoundingClientRect();
	return {X: e.clientX - rect.left, Y: e.clientY - rect.top}
}

//SOCKET FUNCTIONS
socket.onopen = function(e) {
	console.log("socket connection open")
}
socket.onmessage = function(e) {
	let message = JSON.parse(e.data);
	if(message.Type === "ID") {
		id = message.ID
	} else if(message.Type === "STATE") {
		messageInfo.textContent=JSON.stringify(message)
		gameState = message.GameState
		let player = gameState.Players[id]
		playerInfo.textContent=JSON.stringify(player);
		if(player) {
			playerCoords = {X:player.Coords.X, Y:player.Coords.Y}
		}
	}
}


function sendMessage(message) {
	socket.send(JSON.stringify(message));
}

function startGame(){
	let startMessage = {}
	startMessage.Type = "START"
	sendMessage(startMessage)
}

function updateMousePos() {
	let posMessage = {}
	posMessage.Type = "MOUSEPOS"
	posMessage.ID = id
	posMessage.MouseX = mousePos.X + playerCoords.X - windowSize/2
	posMessage.MouseY = mousePos.Y + playerCoords.Y - windowSize/2
	mouseInfo.textContent=JSON.stringify(posMessage);
	sendMessage(posMessage)
}

//DRAW FUNCTIONS
function render() {
	ctx.fillStyle = 'green'
	ctx.fillRect(0,0,windowSize, windowSize)
	ctx.fillStyle = 'red'
	let p = gameState.Players
	Object.keys(p).forEach(key => {
		let midX = p[key].Coords.X - playerCoords.X + 10
		let midY = p[key].Coords.Y - playerCoords.Y + 10
		ctx.fillRect(midX-10, midY-10, midX+10, midY+10)
	})
}

const startButton = document.getElementById("start-button");
startButton.onclick = async function(){
	startGame()
	while(true){
		render()
		updateMousePos()
		await new Promise(r => setTimeout(r, 33)) 
	}
}
