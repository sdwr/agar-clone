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

let canvasSize = 800;
resizeCanvas()
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
window.addEventListener('resize', resizeCanvas, false)
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
		gameState = message.Payload.State
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
	resizeCanvas()
}

//HELPER FUNCTIONS

//note player coords are in game space
function convertScreenToGameCoords(pos) {
	let descaledX = (pos.X - canvas.width/2) / calculateScale()
	let descaledY = (pos.Y - canvas.height/2) / calculateScale()
	return {X: descaledX + playerCoords.X,
		Y: descaledY + playerCoords.Y}
}

function convertGameCoordsToScreen(pos) {
	let scaledX = (pos.X - playerCoords.X) * calculateScale()
	let scaledY = (pos.Y - playerCoords.Y) * calculateScale()
	return {X: scaledX + canvas.width/2,
		Y: scaledY + canvas.height/2}
}

function calculateScale(){
		return 50 / player.Size
}

function updateMousePos() {
	let posMessage = {}
	posMessage.Type = "MOUSEPOS"
	posMessage.ID = id
	posMessage.Payload = {}
	let gameCoords = convertScreenToGameCoords(mousePos)
	posMessage.Payload.MouseX = gameCoords.X
	posMessage.Payload.MouseY = gameCoords.Y 
	mouseInfo.textContent=JSON.stringify(posMessage);
	sendMessage(posMessage)
}

//DRAW FUNCTIONS

function resizeCanvas() {
    canvas.width = window.innerWidth*.9
    canvas.height = window.innerHeight*.9
}
function render() {
	//map boundary background doesnt have to be scaled
	ctx.fillStyle = 'black'
	ctx.fillRect(0,0,canvas.width,canvas.height)

	//background does
	drawRectFromGameCoords({X:0,Y:0},{X:gameState.Size,Y:gameState.Size},"#a5ff8f")
	
	let obs = gameState.Objects
	if(obs) {
	    Object.keys(obs).forEach(key => {
	        let o = obs[key]
		drawObject(o)
	    })
	}
	//death screen also doesn't need to be scaled
	if(player && player.RespawnMillis > 0) {
		drawDeathScreen()
	}
}

function drawRectFromGameCoords(pos, pos2, color) {
	let s1 = convertGameCoordsToScreen(pos)
	let s2 = convertGameCoordsToScreen(pos2)
	ctx.fillStyle = color
	ctx.fillRect(s1.X, s1.Y, s2.X-s1.X, s2.Y-s1.Y)
}

function drawRectFromGameMidpoint(pos, size, color) {
	let halfSize = size/2
	
	drawRectFromGameCoords({X:pos.X-halfSize,Y:pos.Y-halfSize}
		,{X:pos.X+halfSize,Y:pos.Y+halfSize}, color)
}

function drawDeathScreen() {
	ctx.fillStyle = "rgba(125,125,125,0.4)"
	ctx.fillRect(0,0,canvas.width, canvas.height)
}

function drawObject(o) {
	let color = 'green'
	if(o.Type == "PELLET") {
		color = 'yellow'	
	} else if(o.Type == "PLAYER") {
		color = 'red'
	}
	if(o.ID == player.ID) {
		color = "green"
	}
	if(o.Type == "PLAYER" && o.RespawnMillis > 0) {
		return
	}
	drawRectFromGameMidpoint(o.Coords,o.Size,color)
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
