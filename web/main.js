//GLOBAL CONSTANTS
const SERVER_URL = "http://localhost:4404"

const canvas = document.getElementById("game-canvas");

const socket = new Websocket("ws://"+SERVER_URL+"/socket");

socket.onopen = function(e) {
	alert("socket connection open")
	socket.send

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
	posMessage.MouseX = 0
	posMessage.MouseY = 0
}
