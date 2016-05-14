(function e(t,n,r){function s(o,u){if(!n[o]){if(!t[o]){var a=typeof require=="function"&&require;if(!u&&a)return a(o,!0);if(i)return i(o,!0);var f=new Error("Cannot find module '"+o+"'");throw f.code="MODULE_NOT_FOUND",f}var l=n[o]={exports:{}};t[o][0].call(l.exports,function(e){var n=t[o][1][e];return s(n?n:e)},l,l.exports,e,t,n,r)}return n[o].exports}var i=typeof require=="function"&&require;for(var o=0;o<r.length;o++)s(r[o]);return s})({1:[function(require,module,exports){
var IRC = require('../irc');

var MessageList = React.createClass({
	displayName: "MessageList",

	propTypes: {
		messages: React.PropTypes.arrayOf(React.PropTypes.instanceOf(IRC.Message)).isRequired
	},

	componentWillUpdate: function () {
		//Determine if we're at the bottom of the message list
		var node = ReactDOM.findDOMNode(this);
		this.atBottom = node.scrollTop + node.offsetHeight >= node.scrollHeight;
	},

	componentDidUpdate: function () {
		if (this.atBottom) {
			//If we're at the bottom, make sure we stay at the bottom
			var node = ReactDOM.findDOMNode(this);
			node.scrollTop = node.scrollHeight;
		}
	},

	render: function () {
		var rows = [];
		for (var k = 0; k < this.props.messages.length; k++) {
			rows.push(React.createElement("span", { key: k }, this.props.messages[k].Nick() === null ? "You: " : this.props.messages[k].Nick(), this.props.messages[k].DisplayText()));
		}
		return React.createElement("div", { className: "messagelist col-xs-10" }, rows);
	}
});

module.exports = MessageList;

},{"../irc":7}],2:[function(require,module,exports){
var NickList = React.createClass({
	displayName: "NickList",

	propTypes: {
		users: React.PropTypes.arrayOf(React.PropTypes.string).isRequired
	},

	render: function () {
		var rows = [];
		for (var k = 0; k < this.props.users.length; k++) rows.push(React.createElement("li", { className: "nick", key: k }, this.props.users[k]));
		return React.createElement("div", { className: "nicklist" }, React.createElement("ul", { className: "col-xs-2" }, rows));
	}
});

module.exports = NickList;

},{}],3:[function(require,module,exports){
var MessageList = require('./messageList');
var NickList = require('./nickList');

var IRC = require('../irc');

var Room = React.createClass({
	displayName: 'Room',

	propTypes: {
		name: React.PropTypes.string.isRequired,
		users: React.PropTypes.arrayOf(React.PropTypes.string).isRequired,
		messages: React.PropTypes.arrayOf(React.PropTypes.instanceOf(IRC.Message)).isRequired
	},

	render: function () {
		var isRoom = this.props.name[0] == '#';
		return React.createElement('div', { className: 'row' }, React.createElement(MessageList, { messages: this.props.messages }), isRoom ? React.createElement(NickList, { users: this.props.users }) : null);
	}
});

module.exports = Room;

},{"../irc":7,"./messageList":1,"./nickList":2}],4:[function(require,module,exports){
var Room = require('./room');

var IRC = require('../irc');

var TabbedRooms = React.createClass({
	displayName: 'TabbedRooms',

	propTypes: {
		rooms: React.PropTypes.arrayOf(React.PropTypes.instanceOf(IRC.Room)),
		activeRoom: React.PropTypes.instanceOf(IRC.Room),
		onChange: React.PropTypes.func.isRequired,
		onClose: React.PropTypes.func.isRequired,
		defaultName: React.PropTypes.string.isRequired
	},

	render: function () {
		var self = this;
		var activeRoom = this.props.activeRoom;
		var rooms = this.props.rooms;
		var onChange = this.props.onChange;
		var onClose = this.props.onClose;
		var defaultName = this.props.defaultName;
		return React.createElement('div', null, React.createElement('ul', { className: 'tabs' }, rooms.map(function (room) {
			return React.createElement('li', { className: activeRoom.Name() === room.Name() ? "active" : "",
				onClick: onChange.bind(null, room.Name()),
				key: room.Name() }, room.Name(), room.Name() !== defaultName ? React.createElement('span', { onClick: onClose.bind(null, room.Name()) }, '[X]') : null);
		})), React.createElement(Room, { name: activeRoom.Name(), users: activeRoom.Users(), messages: activeRoom.Messages() }));
	}
});

module.exports = TabbedRooms;

},{"../irc":7,"./room":3}],5:[function(require,module,exports){
var Input = React.createClass({
	displayName: "Input",

	propTypes: {
		onChange: React.PropTypes.func,
		onSend: React.PropTypes.func,
		value: React.PropTypes.string
	},
	render: function () {
		return React.createElement("div", { className: "ircinput row" }, React.createElement("input", { type: "text", value: this.props.value, className: "col-xs-11",
			onKeyDown: this.checkForSend, onChange: this.props.onChange }), React.createElement("button", { className: "col-xs-1", onClick: this.props.onSend }, "Send"));
	},

	checkForSend: function (event) {
		if (event.key == 'Enter') this.props.onSend(event);
	}
});

module.exports = Input;

},{}],6:[function(require,module,exports){
(function (global){
//index.js
var React = typeof window !== "undefined" ? window['React'] : typeof global !== "undefined" ? global['React'] : null;
var ReactDOM = typeof window !== "undefined" ? window['ReactDOM'] : typeof global !== "undefined" ? global['ReactDOM'] : null;

var IRCStore = require('./ircstore');
var TabbedRooms = require('./components/tabbedRooms');
var Input = require('./components/textInput');

//The react interface for the IRC client.
var IRCWebChat = React.createClass({
	displayName: 'IRCWebChat',

	getInitialState: function () {
		return {
			rooms: IRCStore.Rooms(),
			activeRoom: IRCStore.Room(IRCStore.DefaultChannel),
			input: { value: "" }
		};
	},

	//Start the connection when the client mounts.
	componentWillMount: function () {
		IRCStore.AddChangeListener(this._updateIRCState);
		IRCStore.Start(window.location.host + "/chat/socket");
	},

	_updateIRCState: function () {
		var rooms = IRCStore.Rooms();
		var activeRoom = IRCStore.ActiveRoom();
		this.setState({ rooms: rooms, activeRoom: activeRoom });
	},

	//sendMessage is called when the user hits enter or click send.
	//It tells the IRCStore to send the message.
	sendMessage: function (event) {
		event.preventDefault();

		var val = this.state.input.value;
		if (val.length > 0 && val[0] == '/') val = val.substring(1, val.length);else if (this.state.activeRoom !== undefined && this.state.activeRoom.Name() != IRCStore.DefaultChannel) {
			val = "PRIVMSG " + this.state.activeRoom.Name() + " :" + val;
		}

		IRCStore.SendMessage(val);
		this.setState({ input: { value: '' } });
	},

	//Listens for the user switching tabs
	_tabChanged: function (newValue) {
		IRCStore.SetActiveRoom(newValue);
	},

	//Kustebs for the close tab button being hit
	_closeTab: function (roomName) {
		if (roomName === IRCStore.DefaultChannel) return;
		IRCStore.CloseRoom(roomName);
	},

	//Listens for changes to the Input box
	_inputChange: function (event) {
		this.setState({ input: { value: event.target.value } });
	},

	render: function () {
		return React.createElement('div', { className: 'container-fluid' }, React.createElement(TabbedRooms, { rooms: this.state.rooms, activeRoom: this.state.activeRoom, onChange: this._tabChanged, onClose: this._closeTab, defaultName: IRCStore.DefaultChannel }), React.createElement(Input, { value: this.state.input.value, onChange: this._inputChange, onSend: this.sendMessage }));
	}
});

ReactDOM.render(React.createElement(IRCWebChat, null), document.getElementById('ircwebchat'));

}).call(this,typeof global !== "undefined" ? global : typeof self !== "undefined" ? self : typeof window !== "undefined" ? window : {})
},{"./components/tabbedRooms":4,"./components/textInput":5,"./ircstore":8}],7:[function(require,module,exports){
//Helper methods to parse irc messages.
//[9:fullstring, 1: nick, 2: user, 3: host]
var userRegex = /(\S+)!(\S+)@(\S+)/;

//Message represents an IRC message. The constructor parses the message
//and provides accessor methods to the various fields of a message
class Message {
    constructor(msg) {
        this.message = msg;
        var rval = parseMessage(msg);
        this.prefix = rval.prefix;
        this.nick = rval.nick;
        this.user = rval.user;
        this.host = rval.host;
        this.command = rval.command.toUpperCase();
        this.args = rval.args;

        this._setDisplayText();
    }

    Prefix() {
        return this.prefix;
    }

    Nick() {
        return this.nick;
    }

    User() {
        return this.user;
    }

    Host() {
        return this.host;
    }

    Command() {
        return this.command;
    }

    Args(index) {
        if (index === undefined) {
            return this.args;
        }
        return this.args[index];
    }

    toString() {
        return this.message;
    }

    DisplayText() {
        return this.displayText;
    }

    _setDisplayText() {
        var cmd = this.Command();
        if (cmd === "PRIVMSG") {
            var msg = this.Args(1);
            if (msg.startsWith(" ACTION")) {
                this.displayText = msg.substring(7);
            } else {
                this.displayText = ": " + msg;
            }
        } else if (cmd === "JOIN") {
            this.displayText = " joined the room.";
        } else if (cmd === "PART") {
            this.displayText = " has left the room.";
        } else if (cmd === "QUIT") {
            this.displayText = " has quit: " + this.Args(0);
        } else {
            this.displayText = this.toString();
        }
    }
}

//Room represents an IRC channel and message queue.
class Room {
    constructor(name) {
        this.name = name;
        this.users = {};
        this._usersArr = [];
        this.messages = [];
    }

    AddMessage(msg) {
        this.messages.push(msg);
    }

    Name() {
        return this.name;
    }

    Messages() {
        return this.messages;
    }

    Users() {
        return this._usersArr;
    }

    AddUser(...users) {
        //TODO: Add users more efficiently
        for (var i in users) {
            var user = users[i];
            var name = user;
            if (user[0] == "@" || user[0] == "+") {
                name = user.substring(1);
            }
            console.log("Adding user: ", user);

            this.users[name] = user;
        }
        this._updateUserArr();
    }

    RemoveUser(...users) {
        for (var i in users) {
            var user = users[i];
            var name = user;
            if (user[0] == "@" || user[0] == "+") {
                name = user.substring(1);
            }
            delete this.users[name];
        }
        this._updateUserArr();
    }

    ClearUsers() {
        this.users = {};
        this._usersArr = [];
    }

    _updateUserArr() {
        this._usersArr = [];
        for (var key in this.users) {
            this._usersArr.push(this.users[key]);
        }
    }
}

//Helper function to parse a raw irc line
function parseMessage(message) {
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
    if (s[0] == ":") {
        var end = s.indexOf(' ');
        retval.prefix = s.substring(1, s.indexOf(' '));

        var prefixArr = retval.prefix.match(userRegex);

        if (prefixArr != null && prefixArr.length >= 4) {
            retval.nick = prefixArr[1];
            retval.user = prefixArr[2];
            retval.host = prefixArr[3];
        }

        s = s.substring(end + 1, s.length);
    }

    //Parse the command
    var end = s.indexOf(' ');
    if (end < 0) //No arg commands will have no space
        end = s.length;
    retval.command = s.substring(0, end).toUpperCase();

    //Parse the parameters by white space, everything after the ':' treated as one argument
    s = s.substring(end + 1, s.length);
    for (; s.length > 0;) {
        if (s[0] == ':') {
            retval.args.push(s.substring(1, s.length));
            break;
        }

        end = s.indexOf(' ');
        if (end < 0) {
            if (s.length > 0) retval.args.push(s);
            break;
        } else {
            retval.args.push(s.substring(0, end));
            if (end + 1 >= s.length) break;
            s = s.substring(end + 1, s.length);
        }
    }

    return retval;
}

//Helper function to parse the prefix of a raw irc line
function parsePrefix(prefix) {
    var prefixarray = prefix.match(userRegex);

    if (prefixarray != null && prefixarray.length > 3) return {
        prefix: prefixarray[0],
        nick: prefixarray[1],
        user: prefixarray[2],
        host: prefixarray[3]
    };
    return null;
}

var IRC = {
    Room: Room,
    Message: Message
};

module.exports = IRC;

},{}],8:[function(require,module,exports){
'use strict';

var IRC = require("./irc");

var _callbacks = []; //Array of callbacks
var websocket;

var SERVER_CH = "Server Messages";
/* The IRCStore is the interface between the react components, and the actual datastructures
	that communicate with the server and manage the client state.

*/
class IRCStore {
	constructor() {
		this.websocket = undefined;
		this.roomsMgr = new RoomsManager();
		this.prevActiveRoom = SERVER_CH;
		this.activeRoom = SERVER_CH;
		this._callbacks = [];
	}

	//Registers a change listener.
	AddChangeListener(callback) {
		this._callbacks.push(callback);
	}

	//Create a new websocket at the provided address.
	Start(wsaddr) {
		var protocol = window.location.protocol === "https:" ? "wss://" : "ws://";
		this.websocket = new WebSocket(protocol + wsaddr);
		this.websocket.onmessage = this._recieveMessage.bind(this);
		this.websocket.onclose = this._socketClose.bind(this);
		var websocket = this.websocket;
		//Send sessionid over ws:
		this.websocket.onopen = function () {
			var sessionID = getCookieValue("SessionID");
			console.log("Session ID: " + sessionID);
			websocket.send(sessionID + "\r\n");
		};
	}

	//Returns the Room object that is active
	ActiveRoom() {
		if (this.roomsMgr.RoomExists(this.activeRoom)) {
			return this.roomsMgr.Room(this.activeRoom);
		}
		return this.roomsMgr.Room(SERVER_CH);
	}

	//Sets the activeRoom name
	SetActiveRoom(rmName) {
		this.prevActiveRoom = this.activeRoom;
		this.activeRoom = rmName;
		this._updateCallbacks();
	}

	SendMessage(msg) {
		//TODO: Parse message depending on context
		this.websocket.send(msg.trim() + "\r\n");
		this.roomsMgr.AddMessage(new IRC.Message(msg.trim()));
		this._updateCallbacks();
	}

	Rooms() {
		return this.roomsMgr.Rooms();
	}

	Room(rmName) {
		return this.roomsMgr.Room(rmName);
	}

	CloseRoom(rmName) {
		this.roomsMgr.RemoveRoom(rmName);
		this.activeRoom = this.prevActiveRoom;
		this.SendMessage("PART " + rmName);
		//this._updateCallbacks();
	}

	_recieveMessage(e) {
		this.roomsMgr.AddMessage(new IRC.Message(e.data.trim()));
		this._updateCallbacks(this.roomsMgr.Rooms());
	}

	_socketClose(e) {
		console.log("Socket closed: ", e);
		this.roomsMgr.AddMessage(new IRC.Message("Websocket to webserver has closed."));
		this._updateCallbacks(this.roomsMgr.Rooms());
	}

	_updateCallbacks() {
		for (var k = 0; k < this._callbacks.length; k++) {
			this._callbacks[k]();
		}
	}
}

class RoomsManager {
	constructor() {
		this.mynick = undefined;
		this.rooms = {};
		this.namesCommand = false; //True if /names w/o args was sent to server
		this.roomGettingUpdates = []; //Tracks 353/366 commands
		this._createRoom(SERVER_CH);
	}

	//Adds a message to the rooms manager, creating a room if it does not exist
	AddMessage(message) {
		if (message.Command() === "NICK") {
			if (message.Prefix() === null) {
				this.mynick = message.Args(0);
			} else if (message.Args().length >= 1) {
				this._changeNick(message.Nick(), message.Args(0));
			}
			return;
		}

		if (message.Command() === "PRIVMSG") {
			var room = message.Args(0);
			if (room == this.mynick && message.Nick() !== null) {
				room = message.Nick();
			}

			if (!this.RoomExists(room)) {
				this._createRoom(room);
			}

			this.Room(room).AddMessage(message);

			return;
		}

		if (message.Command() === "JOIN") {
			console.log("JOIN command...");
			var room = message.Args(0);
			if (room === undefined) return; //Malformed JOIN request

			if (!this.RoomExists(room)) this._createRoom(room);
			if (message.Nick() !== null) {
				this.Room(room).AddUser(message.Nick());
			}
			this.Room(room).AddMessage(message);
			return;
		}

		if (message.Command() === "PART") {
			var room = message.Args(0);
			var user = message.Nick() || this.mynick;
			if (room === undefined) return; //Malformed PART request

			if (message.Nick() === null) {
				//User parting channel
				this.RemoveRoom(room);
			} else if (this.RoomExists(room)) {
				this.Room(room).RemoveUser(user);
			}
			if (this.RoomExists(room)) {
				this.Room(room).AddMessage(message);
			}
			return;
		}

		if (message.Command() === "NAMES") {
			if (message.Args().length <= 0) {
				this.namesCommand = true;
			}
		}

		if (message.Command() === "353") {
			//353 command tells client what users are in a channel,
			//or may be part of a list of all public channels

			if (this.namesCommand) {
				//TODO: Server is sending a list of all public channels. We should show this to the user
				return;
			}

			//Must be a list of users in a specific channel
			//:tepper.freenode.net 353 nick @ #gotest :goirctest @Oooska
			var room = message.Args(2);
			var users = message.Args(3);

			if (room === undefined || users === undefined) {
				console.log("Recieved malformed 353 request");
				return; //Malformed 353 command		
			}

			console.log("Expecting user info for: ", this.roomGettingUpdates);
			if (this.roomGettingUpdates.indexOf(room) <= 0) {
				this.roomGettingUpdates.push(room);
				this.Room(room).ClearUsers();
			}

			users = users.split(" ");
			this._addUser(room, ...users);

			return;
		}

		if (message.Command() === "366") {
			//363 command tells client we're done updating names list
			//:tepper.freenode.net 366 goirctest #gotest :End of /NAMES list.
			if (this.namesCommand) {
				//Done sending a list of all public channels
				this.namesCommand = false;
				return;
			}

			var room = message.Args(1);
			if (room !== undefined) {
				var i = this.roomGettingUpdates.indexOf(room);
				if (i >= 0) {
					this.roomGettingUpdates.splice(i, 1);
				}
			}

			return;
		}

		//Else - some other message. Send it to SERVER_CH to notify user
		this.rooms[SERVER_CH].AddMessage(message);
	}

	_createRoom(name) {
		console.log("_createRoom(", name, ") called.");
		if (this.RoomExists(name)) return;
		console.log("Does not exist... creating");
		this.rooms[name] = new IRC.Room(name);
	}

	RemoveRoom(name) {
		delete this.rooms[name];
	}

	_addUser(roomName, ...user) {
		var room = this.rooms[roomName];
		if (room !== undefined) {
			room.AddUser(...user);
		}
	}

	RoomExists(name) {
		return this.rooms[name] !== undefined;
	}

	Room(room) {
		return this.rooms[room];
	}

	//Returns a list of Rooms
	Rooms() {
		var arr = [];
		var keys = Object.keys(this.rooms);
		for (var k = 0; k < keys.length; k++) arr.push(this.rooms[keys[k]]);

		return arr;
	}

	_changeNick(oldnick, newnick) {
		console.log("TODO: ", oldnick, " changed their name to ", newnick);
	}

	_addPrivMessage(message) {
		//:nick PRIVMSG #channel :Message... (from remoteuser to channel)
		//:nick PRIVMSG user :Message... (from remoteuser to user)
		//PRIVMSG #channel/user :Message (from user to channel/remoteuser)
		var roomName;
		if (message.Nick() !== undefined) //Coming from someone else - roomname is either channel or user that sent it
			roomName = message.Args(0) === this.mynick ? message.Nick() : message.Args(0);else //Outgoing message from our user
			roomName = message.Args(0);

		if (roomName === undefined) return; //Invalid privmsg

		if (!this.RoomExists(roomName)) this._createRoom(roomName);

		this.rooms[roomName].AddMessage(message);
	}
}

//Helper function that returns the value of the specified cookie name
function getCookieValue(name) {
	var value = "; " + document.cookie;
	var parts = value.split("; " + name + "=");
	if (parts.length == 2) return parts.pop().split(";").shift();
}

var store = new IRCStore();
store.DefaultChannel = SERVER_CH;
module.exports = store;

},{"./irc":7}]},{},[6]);
