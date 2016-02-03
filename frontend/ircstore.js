'use strict';


var _callbacks = []; //Array of callbacks
var websocket;

var IRCStore = {
	//Registers a change listener. 
	addChangeListener: function(callback){
		_callbacks.push(callback);
	},


	start: function(wsaddr){
        websocket = new WebSocket("ws://"+wsaddr);
        websocket.onmessage = this._recieveMessage;
        websocket.onclose = this._socketClose;
	},

	sendMessage: function(msg){
		//TODO: Parse message depending on context
		websocket.send(msg+"\r\n");
		Rooms.addMessage(msg);
		updateCallbacks(Rooms.asArray());
	},

	_recieveMessage: function(e){
		Rooms.addMessage(e.data);
		updateCallbacks(Rooms.asArray());
	},

	_socketClose: function(e){
		console.log("Socket closed: ", e)
		Rooms.addMessage("Websocket to webserver has closed.")
		updateCallbacks(Rooms.asArray());
	},


}

function updateCallbacks(rooms){
	for(var k=0; k < _callbacks.length; k++)
		_callbacks[k](rooms)
}


//Data structure to hold the rooms, nick list, etc.
//TODO: Build a proper ES6 class to manage this. 
var Rooms = {
	rooms: { "server": {name : "server", users: ["irc server"], messages: []}},

	asArray: function(){
		var arr = [];
		var keys = Object.keys(this.rooms);
		for(var k = 0; k < keys.length; k++)
			arr.push(this.rooms[keys[k]]);

		return arr;
	},

	addRoom: function(room, users) {
		if(this.rooms[room] !== undefined)
			return;

		if(users === null || users === undefined)
			users = [];

		this.rooms[room] = {
			name: room, 
			users: users, 
			messages: []
		};
	}, 

	removeRoom: function(room){
		this.rooms[room] = undefined;
	},

	addMessageToRoom: function(room, message){
		this.rooms[room].messages.push(message);	
	},

	addMessage: function(rawmessage){
		rawmessage = rawmessage.trim();
		var pMessage = parseMessage(rawmessage);

		var room = "server";
		var output = rawmessage;
		if(pMessage.command == "PRIVMSG" && pMessage.args.length >= 2){
			room = pMessage.args[0];
			output = pMessage.nick +": "+pMessage.args[1];
		} else if (pMessage.command == "JOIN" && pMessage.args.length >= 1){
			room = pMessage.args[0];
			this.addUser(room, pMessage.nick);
			output = ">>> "+pMessage.nick+" has joined the channel.";
		} else if(pMessage.command == "PART" && pMessage.args.length >= 1){
			room = pMessage.args[0];
			this.removeUser(room, pMessage.nick);
			output = "<<< "+pMessage.nick+" has left the channel.";
		}

		if(this.rooms[room] === undefined)
			this.addRoom(room);

		this.addMessageToRoom(room, output);
	},

	addUser: function(room, user){
		if(this.rooms[room] === undefined)
			return;
		this.rooms[room].users.push(user);
	},

	removeUser: function(room, user){
		//TODO
	}

}


//Helper methods to parse irc messages.

//[0:fullstring, 1: prefix, 2: command, 3: destination, 4: message contents]
var ircRegex = /:(\S+) (\S+) (\S+) ([:print:]+)/

//[9:fullstring, 1: nick, 2: user, 3: host]
var userRegex = /(\S+)!~(\S+)@(\S+)/

function parseMessage(message){
	var retval = {
		prefix: null, //nick!~user@host
		nick: null,
		user: null,
		host: null,
		command: null, //PRIVMSG or other command
		args: [] //Argument for command
	};

	//Parse the prefix if it is present (:Oooska1!~Oooska@knds.xdsl.dyn.ottcommunications.com)
	var s = message;
	if(s[0] == ":"){
		var end = s.indexOf(' ');
		retval.prefix = s.substring(1, s.indexOf(' '));

		var prefixArr = retval.prefix.match(userRegex);

		if(prefixArr != null && prefixArr.length >= 4){
			retval.nick = prefixArr[1];
			retval.user = prefixArr[2];
			retval.host = prefixArr[3];
		}

		s = s.substring(end+1, s.length);
	}


	//Parse the command
	var end = s.indexOf(' ');
	retval.command = s.substring(0, end);


	//Parse the parameters by whhite space, everything after the ':' treated as one argument
	s = s.substring(end+1, s.length);
	for (;s.length > 0;){
		console.log("Test...")
		if(s[0] == ':'){
			retval.args.push(s.substring(1,s.length));
			break;
		}

		end = s.indexOf(' ');
		if(end < 0){
			if(s.length > 0)
				retval.args.push(s);
			break;
		} else {
			retval.args.push(s.substring(0, end));
			if(end+1 >= s.length)
				break;
			s = s.substring(end+1, s.length);
		}

	}

	console.log("Parsed message: ", retval)
	return retval;
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

module.exports = IRCStore;