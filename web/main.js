console.log("file loaded")
//GLOBAL CONSTANTS
const SERVER_URL = "localhost:4404"

const canvas = document.getElementById("game-canvas");
const ctx = canvas.getContext('2d');
const socket = new WebSocket("ws://"+SERVER_URL+"/socket");

const gameState = {};
const id = "";
const windowSize = 800;

//DEFINITIONS
class Position {
	constructor(x, y) {
		this.x = x;
		this.y = y;
	}

}

//SOCKET FUNCTIONS
socket.onopen = function(e) {
	console.log("socket connection open")
}
socket.onmessage = function(e) {
	let message = JSON.parse(e.data);
	console.log(message)
	if(message.type === "ID") {
		id = message.Sender
	} else if(message.type === "STATE") {
		gameState = message.GameState
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
	posMessage.mouseX = 0
	posMessage.mouseY = 0
	sendMessage(posMessage)
}

//DRAW FUNCTIONS
function render() {
	console.log(gameState);
}

const startButton = document.getElementById("start-button");
startButton.onclick = async function(){
	startGame()
	while(true){
		render()
		await new Promise(r => setTimeout(r, 33)) 
	}
}
