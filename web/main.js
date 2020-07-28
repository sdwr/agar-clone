//GLOBAL CONSTANTS
const SERVER_URL = "http://localhost:4404"

const canvas = document.getElementById("game-canvas");
const ctx = canvas.getContext('2d');
const socket = new Websocket("ws://"+SERVER_URL+"/socket");

const gameState = {};

socket.onopen = function(e) {
	alert("socket connection open")
}
socket.onmessage = function(e) {
	gameState = JSON.parse(e.data);
}


function sendMessage(message) {
	socket.send(JSON.stringify(message));
}

function startGame(){
	startMessage = {}
	startMessage.Type = "START"
	sendMessage(startMessage)
}

function updateMousePos() {
	posMessage = {}
	posMessage.Type = "MOUSEPOS"
	posMessage.mouseX = 0
	posMessage.mouseY = 0
	sendMessage(posMessage)
}

function render() {
	console.log(gameState);
}

const startButton = document.getElementById("start-button");
startButton.onclick = async function(){
	startGame()
	while(true){
		render()
		await sleep(20);
	}
}
