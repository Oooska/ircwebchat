'use strict';
var IRC = require("./irc")

var _callbacks = []; //Array of callbacks
var websocket;

var SERVER_CH = "Server Messages";
/* The IRCStore is the interface between the react components, and the actual datastructures
	that communicate with the server and manage the client state.

*/
class IRCStore {
	constructor(){
		this.websocket = undefined;
		this.roomsMgr = new RoomsManager();
		this.prevActiveRoom = SERVER_CH;
		this.activeRoom = SERVER_CH;
		this._callbacks = [];
	}
	
	//Registers a change listener. 
	AddChangeListener(callback){
		this._callbacks.push(callback);
	}

	//Create a new websocket at the provided address.
	Start(wsaddr){
		var protocol = (window.location.protocol === "https:" ? "wss://" : "ws://")
        this.websocket = new WebSocket(protocol+wsaddr);
        this.websocket.onmessage = this._recieveMessage.bind(this);
        this.websocket.onclose = this._socketClose.bind(this);
		var websocket = this.websocket;
		//Send sessionid over ws:
		this.websocket.onopen = function(){
			var sessionID = getCookieValue("SessionID")
			console.log("Session ID: "+sessionID)
			websocket.send(sessionID+"\r\n")
		};
	}

	//Returns the Room object that is active
	ActiveRoom(){
		if(this.roomsMgr.RoomExists(this.activeRoom)){
			return this.roomsMgr.Room(this.activeRoom);
		}
		return this.roomsMgr.Room(SERVER_CH);
	}
	
	//Sets the activeRoom name
	SetActiveRoom(rmName){
		this.prevActiveRoom = this.activeRoom;
		this.activeRoom = rmName;
		this._updateCallbacks()
	}

	SendMessage(msg){
		//TODO: Parse message depending on context
		this.websocket.send(msg.trim()+"\r\n");
		this.roomsMgr.AddMessage(new IRC.Message(msg.trim()));
		this._updateCallbacks();
	}

	Rooms(){
		return this.roomsMgr.Rooms();
	}
	
	Room(rmName){
		return this.roomsMgr.Room(rmName);
	}

	CloseRoom(rmName){
		this.roomsMgr.RemoveRoom(rmName);
		this.activeRoom = this.prevActiveRoom;
		this.SendMessage("PART "+rmName);
		//this._updateCallbacks();
	}

	_recieveMessage(e){
		this.roomsMgr.AddMessage(new IRC.Message(e.data.trim()));
		this._updateCallbacks(this.roomsMgr.Rooms());
	}

	_socketClose(e){
		console.log("Socket closed: ", e)
		this.roomsMgr.AddMessage(new IRC.Message("Websocket to webserver has closed."))
		this._updateCallbacks(this.roomsMgr.Rooms());
	}

	_updateCallbacks(){
		for(var k=0; k < this._callbacks.length; k++){
			this._callbacks[k]();
		}
	}
}

class RoomsManager {
	constructor(){
		this.mynick = undefined;
		this.rooms = {};
		this.namesCommand = false; //True if /names w/o args was sent to server
		this.roomGettingUpdates = []; //Tracks 353/366 commands
		this._createRoom(SERVER_CH);
	}
	
	//Adds a message to the rooms manager, creating a room if it does not exist
	AddMessage(message){
		if(message.Command() === "NICK"){
			if(message.Prefix() === null){
				this.mynick = message.Args(0);
			} else if(message.Args().length >= 1){
				this._changeNick(message.Nick(), message.Args(0));
			}
			return;
		}
		
		if(message.Command() === "PRIVMSG"){
			var room = message.Args(0);
			if(room == this.mynick && message.Nick() !== null){
				room = message.Nick();
			}
			
			if(!this.RoomExists(room)){
					this._createRoom(room);
			}

			this.Room(room).AddMessage(message);

			return;
		}
		
		if(message.Command() === "JOIN"){
			console.log("JOIN command...")
			var room = message.Args(0);
			if(room === undefined)
				return; //Malformed JOIN request
			
			if(!this.RoomExists(room)) 
				this._createRoom(room);
			if(message.Nick() !== null){
				this.Room(room).AddUser(message.Nick());
			} 
			this.Room(room).AddMessage(message);
			return;
		}
		
		if(message.Command() === "PART"){
			var room = message.Args(0)
			var user = message.Nick() || this.mynick;
			if(room === undefined)
				return; //Malformed PART request
				
			if(message.Nick() === null){
				//User parting channel
				this.RemoveRoom(room);
			} else if(this.RoomExists(room)){
				this.Room(room).RemoveUser(user);
			}
			if(this.RoomExists(room)){
				this.Room(room).AddMessage(message);
			}
			return;
		}
		
		if(message.Command() === "NAMES"){
			if(message.Args().length <= 0){
				this.namesCommand = true;
			} 
		}
		
		if(message.Command() === "353"){
			//353 command tells client what users are in a channel,
			//or may be part of a list of all public channels
			
			if(this.namesCommand){
				//TODO: Server is sending a list of all public channels. We should show this to the user
				return; 
			}
			
			//Must be a list of users in a specific channel
			//:tepper.freenode.net 353 nick @ #gotest :goirctest @Oooska
			var room = message.Args(2);
			var users = message.Args(3);
			
			if(room === undefined || users === undefined){
				console.log("Recieved malformed 353 request");
				return; //Malformed 353 command		
			}

			console.log("Expecting user info for: ", this.roomGettingUpdates)		
			if(this.roomGettingUpdates.indexOf(room) <= 0){
				this.roomGettingUpdates.push(room);
				this.Room(room).ClearUsers();
			}
			
			users = users.split(" ");
			this._addUser(room, ...users);
			
			return;
		}
		
		if(message.Command() === "366"){
			//363 command tells client we're done updating names list
			//:tepper.freenode.net 366 goirctest #gotest :End of /NAMES list.
			if(this.namesCommand){
				//Done sending a list of all public channels
				this.namesCommand = false;
				return;
			}
			
			var room = message.Args(1);
			if(room !== undefined){
				var i = this.roomGettingUpdates.indexOf(room);
				if(i >= 0){
					this.roomGettingUpdates.splice(i, 1);
				}
			}
			
			return;
		}
		
		//Else - some other message. Send it to SERVER_CH to notify user
		this.rooms[SERVER_CH].AddMessage(message);
		
	}
	
	_createRoom(name){
		console.log("_createRoom(",name,") called.")
		if(this.RoomExists(name)) 
			return;
		console.log("Does not exist... creating")
		this.rooms[name] = new IRC.Room(name);
	}
	
	RemoveRoom(name){
		delete this.rooms[name];
	}
	
	_addUser(roomName, ...user){
		var room = this.rooms[roomName];
		if(room !== undefined){
			room.AddUser(...user);
		}
	}
	
	RoomExists(name){
		return this.rooms[name] !== undefined;
	}
	
	Room(room){
		return this.rooms[room]
	}
	
	//Returns a list of Rooms
	Rooms(){
		var arr = [];
		var keys = Object.keys(this.rooms);
		for(var k = 0; k < keys.length; k++)
			arr.push(this.rooms[keys[k]]);

		return arr;
	}
	
	_changeNick(oldnick, newnick){
		console.log("TODO: ", oldnick, " changed their name to ", newnick)
		
	}
	
	_addPrivMessage(message){
		//:nick PRIVMSG #channel :Message... (from remoteuser to channel)
		//:nick PRIVMSG user :Message... (from remoteuser to user)
		//PRIVMSG #channel/user :Message (from user to channel/remoteuser)
		var roomName;
		if(message.Nick() !== undefined) //Coming from someone else - roomname is either channel or user that sent it
			roomName = message.Args(0) === this.mynick ? message.Nick() : message.Args(0);
		else //Outgoing message from our user
			roomName = message.Args(0);
		
		if(roomName === undefined)
			return; //Invalid privmsg
		
		if(!this.RoomExists(roomName))
			this._createRoom(roomName);
			
		this.rooms[roomName].AddMessage(message);
	}
}




//Helper function that returns the value of the specified cookie name
function getCookieValue(name) {
  var value = "; " + document.cookie;
  var parts = value.split("; " + name + "=");
  if (parts.length == 2) return parts.pop().split(";").shift();
}


var store = new IRCStore()
store.DefaultChannel = SERVER_CH;
module.exports = store;