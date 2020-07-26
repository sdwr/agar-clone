//GLOBAL CONSTANTS
const SERVER_URL = "http://localhost:4404"

const canvas = document.getElementById("game-canvas");

const socket = new Websocket("ws://"+SERVER_URL+"/socket");


function createMessage(type


socket.onopen = function(e) {
	alert("socket connection open")
	socket.send

}
