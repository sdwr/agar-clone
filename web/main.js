console.log("file loaded")
//GLOBAL CONSTANTS
const SERVER_URL = "localhost:4404"

const canvas = document.getElementById("game-canvas");
const playerInfo = document.getElementById("player-info");
const mouseInfo = document.getElementById("mouse-info");
const messageInfo = document.getElementById("message-info");
const fpsInfo = document.getElementById("fps-info");
const devTools = document.getElementById("dev-tools");
const ctx = canvas.getContext('2d');
const socket = new WebSocket("ws://"+SERVER_URL+"/socket");

const windowSize = 400;

let lastUpdated = 0;

let gameState = {Players:[]};
let id = 1;

let playerCoords = {X:0,Y:0}
let player = null
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
//open dev tools w ` 
document.addEventListener('keydown', function(e) {
	if(e.keyCode == 192) {
        if (devTools.style.display == "none") {
		    devTools.style.display = "block";
	    } else {
		    devTools.style.display = "none"
	}
	}
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
		//messageInfo.textContent=JSON.stringify(message)
		gameState = message.State
		player = gameState.CurrentPlayer
		if(player) {
			playerCoords = {X:player.Coords.X, Y:player.Coords.Y}
		}
		playerInfo.textContent=JSON.stringify(playerCoords) + JSON.stringify(player);
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



function convertScreenToGameCoords(pos) {
	return {X: pos.X + playerCoords.X - windowSize/2,
		Y: pos.Y + playerCoords.Y - windowSize/2}
}

function convertGameCoordsToScreen(pos) {
	return {X: pos.X - playerCoords.X + windowSize/2,
		Y: pos.Y - playerCoords.Y + windowSize/2}
}

function updateMousePos() {
	let posMessage = {}
	posMessage.Type = "MOUSEPOS"
	posMessage.ID = id
	let gameCoords = convertScreenToGameCoords(mousePos)
	posMessage.MouseX = gameCoords.X
	posMessage.MouseY = gameCoords.Y 
	mouseInfo.textContent=JSON.stringify(posMessage);
	sendMessage(posMessage)
}

//DRAW FUNCTIONS
function render() {
	ctx.fillStyle = 'black'
	ctx.fillRect(0,0,windowSize,windowSize)
	ctx.fillStyle = 'green'
	let screenCoords = convertGameCoordsToScreen({X:0,Y:0})
	ctx.fillRect(screenCoords.X, screenCoords.Y, gameState.Size, gameState.Size)
	let obs = gameState.Objects
	Object.keys(obs).forEach(key => {
		let o = obs[key]
		drawObject(o)
	})
	if(player && player.RespawnMillis > 0) {
		drawDeathScreen()
	}
}

function drawDeathScreen() {
	ctx.fillStyle = "rgba(125,125,125,0.4)"
	ctx.fillRect(0,0,windowSize,windowSize)
}

function drawObject(o) {
	let screenCoords = convertGameCoordsToScreen(o.Coords)
	let halfSize = o.Size/2
	let midX = screenCoords.X 
	let midY = screenCoords.Y
	if(o.Type == "PELLET") {
		ctx.fillStyle = 'yellow'	
	} else if(o.Type == "PLAYER") {
		ctx.fillStyle = 'red'
	}
	if(o.ID == player.ID) {
		ctx.fillStyle = "blue"
	}
	ctx.fillRect(midX-halfSize, midY-halfSize, halfSize*2, halfSize*2)

}

const startButton = document.getElementById("start-button");
startButton.onclick = async function(){
	startGame()
	while(true){
		let now = Date.now()
		fpsInfo.textContent = "" + now-lastUpdated + " ms" 
		render()
		updateMousePos()
		lastUpdated = now
		await new Promise(r => setTimeout(r, 10))
	}
}

const toggleDebug = document.getElementById("toggle-debug");
toggleDebug.onclick = function() {
	let debugPanel = document.getElementById("debug-info")
	if (debugPanel.style.display == "none") {
		debugPanel.style.display = "block";
	} else {
		debugPanel.style.display = "none"
	}
}

const addBot = document.getElementById("create-bot");
addBot.onclick = function createBot(){
	let createBotMessage = {}
	createBotMessage.Type = "CREATEBOT"
	sendMessage(createBotMessage)
}
