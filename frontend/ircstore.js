'use strict';


var _callbacks = []; //Array of callbacks
var websocket;


/* The IRCStore is the interface between the react components, and the actual datastructures
	that communicate with the server and manage the client state.

*/
var IRCStore = {
	//Registers a change listener. 
	addChangeListener: function(callback){
		_callbacks.push(callback);
	},


	//Create a new websocket at the provided address.
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




//The Rooms variable holds the data structure that maintains the current client status.
//TODO: Build a proper ES6 class to manage this. 
//TODO: Rename Rooms to something more appropriate.
/*

The addMessage() function will eventually parse the message and act appropriately
to update the state of the irc client. //TODO: Most of the previous sentence.

The data structure mirrors the react props.
//TODO: 
{
	"#roomname1": {
		name : "#roomname1", 
		users: ["user1", "user2"], 
		messages: ["string1", "string2"]
		updating_353: bool //bool representing whether the room is actively getting a new user list from the server
	},

	"#roomname2": {
		name : "#roomname2", 
		users: ["user2", "user3"], 
		messages: ["string11", "string222"]
	},

	"user2": {
		name : "user2", 
		users: ["user2"], 
		messages: ["string1", "string2"]
	},
}

*/
var ERR_ROOM_404 = "The specified room does not exist."

var Rooms = {
	rooms: { "server": {name : "server", users: ["irc server"], messages: []}},

	addMessage: function(rawmessage){
		rawmessage = rawmessage.trim();
		var pMessage = parseMessage(rawmessage);
		console.log("parsed message: ", rawmessage, pMessage);

		var room = "server";
		var output = rawmessage;
		

		if(pMessage.command == "PRIVMSG" && pMessage.args.length >= 2){
			room = pMessage.args[0];
			output = pMessage.nick +": "+pMessage.args[1];
		
		} else if (pMessage.command == "JOIN" && pMessage.args.length >= 1){
			room = pMessage.args[0];
			if(pMessage.prefix == null){ 
				//The user joined a channel
				if(!this.roomExists())
					this.createRoom(room);
			} else { //Someone else joined a channel
				this.addUser(room, pMessage.nick);
				output = ">>> "+pMessage.nick+" has joined the channel.";
			}
		
		} else if(pMessage.command == "PART" && pMessage.args.length >= 1){
			room = pMessage.args[0];
			this.removeUser(room, pMessage.nick);
			output = "<<< "+pMessage.nick+" has left the channel.";
		
		} else if(pMessage.command == "353") { //Response to /names or /join : 
			//:hobana.freenode.net 353 NotSoShadyBloke = #mainehackerclub :NotSoShadyBloke Oooska1 kroker1 +alsochris
			room = pMessage.args[2];
			if(pMessage.args[3]){
				var users = pMessage.args[3].split(" "); 
				var roomObj = this.getRoom(room);
				if(!roomObj.updating_353){
					//Server is providing a fresh list of users, clear out old list
					roomObj.updating_353 = true;
					this.clearUsers(room);
				}
				
				for(var k = 0; k < users.length; k++){
					this.addUser(room, users[k]);
				}
			}
		} else if(pMessage.command == "366"){ //The end of /names
			//Notifying us the users list is up to date.
			room = pMessage.args[1];
			if(room && this.getRoom(room)){
				//We are done updating the user list
				this.getRoom(room).updating_353 = false;
			}
		}

		if(!this.roomExists())
			this.createRoom(room);

		this._addMessageToRoom(room, output);
	},

	asArray: function(){
		var arr = [];
		var keys = Object.keys(this.rooms);
		for(var k = 0; k < keys.length; k++)
			arr.push(this.rooms[keys[k]]);

		return arr;
	},

	roomExists: function(room){
		return this.rooms[room] !== undefined;
	},

	createRoom: function(room, users) {
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

	getRoom: function(room){
		return this.rooms[room];
	},


	addUser: function(room, user){
		if(this.rooms[room] === undefined)
			throw ERR_ROOM_404;

		if(this.rooms[room].users.indexOf(user) < 0)
			this.rooms[room].users.push(user);
	},

	//clearUsers removes all users from the room
	clearUsers: function(room){
		this.rooms[room].users = [];
	},

	//removeUser removes the user from the room.
	removeUser: function(room, user){
		if(this.rooms[room] === undefined)
			throw ERR_ROOM_404;

		var index = this.rooms[room].users.indexOf(user);
		if(index >= 0){
			this.rooms[room].users.splice(index, 1);
		}

	},

	//Adds the specified message to the end of the room's messagelist.
	_addMessageToRoom: function(room, message){
		this.rooms[room].messages.push(message);	
	},

}


//Helper methods to parse irc messages. 
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
	retval.command = s.substring(0, end).toUpperCase();


	//Parse the parameters by whhite space, everything after the ':' treated as one argument
	s = s.substring(end+1, s.length);
	for (;s.length > 0;){
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

	return retval;
}

function parsePrefix(prefix){
	var prefixarray = prefix.match(userRegex);

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