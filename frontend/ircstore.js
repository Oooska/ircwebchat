'use strict';


var _callbacks = []; //Array of callbacks
var websocket;

//Src: http://www.mybuddymichael.com/writings/a-regular-expression-for-irc-messages.html
//[0:fullstring, 1: prefix, 2: command, 3: destination, 4: message contents]
var ircRegex = /^(?:[:](\S+) )?(\S+)(?: (?!:)(.+?))?(?: [:](.+))?$/

//[9:fullstring, 1: nick, 2: user, 3: host]
var userRegex = /^(:\S+)!~(\S+)@(\S+)$/

function parseMessage(message){
	var msgarray = message.match(ircRegex)

	console.log("parseMsg: ", message, "parsed: ", msgarray);

	if(msgarray != null && msgarray.length > 4)
		return {
			message: msgarray[0],
			prefix: msgarray[1],
			command: msgarray[2],
			destination: msgarray[3],
			contents: msgarray[4]
		};

	return null;
}

function parsePrefix(prefix){
	var prefixarray = prefix.match(userRegex);

	console.log("parsePrefix: ", prefix, " parsed :", prefixarray);

	if(prefixarray != null && prefixarray.length > 3)
	return {
		prefix: prefixarray[0],
		nick: prefixarray[1],
		user: prefixarray[2],
		host: prefixarray[3]
	};
	return null;
}

var _rooms = {
	"server": {name : "server", users: ["irc server"], messages: []}
	//"#room1" : {name: "#room1", users: ["user1", "user2", "user3"], messages: ["message1", "message2", "message3"]}, 
	//"#room2" : {name: "#room2", users: ["user1", "user2", "user4"], messages: ["message4", "message5", "message6"]}, 
	//"someguy" : {name: "someguy", users: ["someguy"], messages: ["privmessage1", "privmessage2", "privmessage3"]}
}

function _roomsAsArray(){
	var arr = [];

	var keys = Object.keys(_rooms);
	for(var k = 0; k < keys.length; k++){
		arr.push(_rooms[keys[k]]);
	}

	return arr;
}

function addRoom(room, users){
	if(_rooms[room] !== undefined)
		return;

	if(users === null || users === undefined)
		users = [];

	_rooms[room] = {
		name: room, 
		users: users, 
		messages: []
	};
}

function removeRoom(room){
	_rooms[room] = undefined;
}

function addMessage(room, message){
	if(_rooms[room] === undefined)
		addRoom(room);
	
	_rooms[room].messages.push(message);
}

function addUser(room, user){
	if(_rooms[room] === undefined)
		return;
	_rooms[room].users.push(user);
}

function removeUser(room, user){
	//TODO
}


var IRCStore = {
	//Registers a change listener. 
	addChangeListener: function(callback){
		_callbacks.push(callback);
	},


	init: function(wsaddr){
        websocket = new WebSocket("ws://"+wsaddr);
        websocket.onmessage = this._recieveMessage;
        websocket.onclose = this._socketClose;
	},

	sendMessage: function(msg){
		//TODO: Parse message depending on context
		websocket.send(msg+"\r\n");
		addMessage("server", "YOU :" +msg);
	},

	_recieveMessage: function(e){
		console.log("Received message: ", e);

		//var msg = parseMessage(e.data);

		//if(msg == null){
			addMessage("server", e.data);
		/*}

		else if(msg.command == "PRIVMSG"){
			//Add message for the given room.
			//Add user/room if they don't exist

			//TODO: Handle private messaging between two people correctly.
			var from = parsePrefix(msg.prefix);
			addMessage(msg.destination, from.nick +": "+msg.contents);
			
		}

		else if(msg.command == "JOIN"){
			var from = parsePrefix(msg.prefix);
			addUser(msg.destination, from.nick);
			addMessage(msg.destination, ">>> "+from.nick+" has joined the channel.");
		}

		else if(msg.command == "PART"){
			var from = parsePrefix(msg.prefix);
			removeUser(msg.destination, from.nick);
			addMessage(msg.destination, "<<< "+from.nick+" has left the channel.");
		}


		else {
			addMessage("server", msg.message)
		}*/

		var rooms = _roomsAsArray();
		for(var k=0; k < _callbacks.length; k++){
			_callbacks[k](rooms);
		}
},

	_socketClose: function(e){
		console.log("Socket closed: ", e)
		addMessage("server", "Websocket to webserver has closed.")
		var rooms = _roomsAsArray();
		for(var k=0; k < _callbacks.length; k++){
			_callbacks[k](rooms)
		}
	}
}


module.exports = IRCStore;