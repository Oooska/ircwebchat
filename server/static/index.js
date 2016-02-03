(function e(t,n,r){function s(o,u){if(!n[o]){if(!t[o]){var a=typeof require=="function"&&require;if(!u&&a)return a(o,!0);if(i)return i(o,!0);var f=new Error("Cannot find module '"+o+"'");throw f.code="MODULE_NOT_FOUND",f}var l=n[o]={exports:{}};t[o][0].call(l.exports,function(e){var n=t[o][1][e];return s(n?n:e)},l,l.exports,e,t,n,r)}return n[o].exports}var i=typeof require=="function"&&require;for(var o=0;o<r.length;o++)s(r[o]);return s})({1:[function(require,module,exports){
(function (global){
//index.js
var React = typeof window !== "undefined" ? window['React'] : typeof global !== "undefined" ? global['React'] : null;
var ReactDOM = typeof window !== "undefined" ? window['ReactDOM'] : typeof global !== "undefined" ? global['ReactDOM'] : null;
var Tabs = require('tabs.react');
var IRCStore = require('./ircstore');

var IRCWebChat = React.createClass({
	displayName: 'IRCWebChat',

	getInitialState: function () {
		return {
			rooms: [{ name: "server", users: ["server"], messages: ["Loading..."] }],

			activeTab: "server",
			input: { value: "" }
		};
	},

	componentWillMount: function () {
		IRCStore.addChangeListener(this.addMessage);
		IRCStore.start(window.location.host + "/chat/socket");
	},

	addMessage: function (newRooms) {
		this.setState({ rooms: newRooms });
	},

	inputChange: function (event) {
		this.setState({ input: { value: event.target.value } });
	},

	sendMessage: function (event) {
		event.preventDefault();

		var val = this.state.input.value;
		if (val.length > 0 && val[0] == '/') val = val.substring(1, val.length);else if (this.activeTab != "") val = "PRIVMSG " + this.state.activeTab + " :" + val;

		console.log("Sending message. Input: ", this.state.input.value, " Parsed to :", val);

		IRCStore.sendMessage(val);
		this.setState({ input: { value: '' } });
	},

	_tabChanged: function (newValue) {
		console.log("new tab: ", newValue);
		this.setState({ activeTab: newValue.active });
	},

	render: function () {
		return React.createElement('div', { className: 'container-fluid' }, React.createElement(TabbedRooms, { rooms: this.state.rooms, activeTab: this.state.activeTab, onChange: this._tabChanged }), React.createElement(Input, { value: this.state.input.value, onChange: this.inputChange, onSend: this.sendMessage }));
	}
});

var TabbedRooms = React.createClass({
	displayName: 'TabbedRooms',

	propTypes: {
		rooms: React.PropTypes.arrayOf(React.PropTypes.shape({
			name: React.PropTypes.string.isRequired,
			users: React.PropTypes.arrayOf(React.PropTypes.string).isRequired,
			messages: React.PropTypes.arrayOf(React.PropTypes.string).isRequired
		})),

		activeTab: React.PropTypes.string.isRequired,
		onChange: React.PropTypes.func.isRequired
	},

	render: function () {
		return React.createElement(Tabs, { active: this.props.activeTab, useState: true, propName: 'name', onChange: this.props.onChange }, this.props.rooms.map(function (room) {
			return React.createElement(Room, { name: room.name, users: room.users, messages: room.messages, key: room.name });
		}));
	}
});

var Room = React.createClass({
	displayName: 'Room',

	propTypes: {
		name: React.PropTypes.string.isRequired,
		users: React.PropTypes.arrayOf(React.PropTypes.string).isRequired,
		messages: React.PropTypes.arrayOf(React.PropTypes.string).isRequired
	},

	render: function () {
		var isRoom = this.props.name[0] == '#';
		return React.createElement('div', { className: 'row' }, React.createElement(MessageList, { messages: this.props.messages }), isRoom ? React.createElement(NickList, { users: this.props.users }) : null);
	}
});

var NickList = React.createClass({
	displayName: 'NickList',

	propTypes: {
		users: React.PropTypes.arrayOf(React.PropTypes.string).isRequired
	},

	render: function () {
		var rows = [];
		for (var k = 0; k < this.props.users.length; k++) rows.push(React.createElement('li', { className: 'nick', key: k }, this.props.users[k]));
		return React.createElement('ul', { className: 'nicklist col-xs-2' }, rows);
	}
});

var MessageList = React.createClass({
	displayName: 'MessageList',

	propTypes: {
		messages: React.PropTypes.arrayOf(React.PropTypes.string).isRequired
	},
	render: function () {
		var rows = [];
		for (var k = 0; k < this.props.messages.length; k++) rows.push(React.createElement('span', { key: k }, this.props.messages[k]));
		return React.createElement('div', { className: 'messagelist col-xs-10' }, rows);
	}
});

var Input = React.createClass({
	displayName: 'Input',

	propTypes: {
		onChange: React.PropTypes.func,
		onSend: React.PropTypes.func,
		value: React.PropTypes.string
	},
	render: function () {
		return React.createElement('div', { className: 'row' }, React.createElement('input', { type: 'text', value: this.props.value, className: 'col-xs-11',
			onKeyDown: this.checkForSend, onChange: this.props.onChange }), React.createElement('button', { className: 'col-xs-1', onClick: this.props.onSend }, 'Send'));
	},

	checkForSend: function (event) {
		if (event.key == 'Enter') this.props.onSend(event);
	}
});

ReactDOM.render(React.createElement(IRCWebChat, null), document.getElementById('ircwebchat'));

}).call(this,typeof global !== "undefined" ? global : typeof self !== "undefined" ? self : typeof window !== "undefined" ? window : {})
},{"./ircstore":2,"tabs.react":3}],2:[function(require,module,exports){
'use strict';

var _callbacks = []; //Array of callbacks
var websocket;

var IRCStore = {
	//Registers a change listener.
	addChangeListener: function (callback) {
		_callbacks.push(callback);
	},

	start: function (wsaddr) {
		websocket = new WebSocket("ws://" + wsaddr);
		websocket.onmessage = this._recieveMessage;
		websocket.onclose = this._socketClose;
	},

	sendMessage: function (msg) {
		//TODO: Parse message depending on context
		websocket.send(msg + "\r\n");
		Rooms.addMessage(msg);
		updateCallbacks(Rooms.asArray());
	},

	_recieveMessage: function (e) {
		Rooms.addMessage(e.data);
		updateCallbacks(Rooms.asArray());
	},

	_socketClose: function (e) {
		console.log("Socket closed: ", e);
		Rooms.addMessage("Websocket to webserver has closed.");
		updateCallbacks(Rooms.asArray());
	}

};

function updateCallbacks(rooms) {
	for (var k = 0; k < _callbacks.length; k++) _callbacks[k](rooms);
}

//Data structure to hold the rooms, nick list, etc.
//TODO: Build a proper ES6 class to manage this.
var Rooms = {
	rooms: { "server": { name: "server", users: ["irc server"], messages: [] } },

	asArray: function () {
		var arr = [];
		var keys = Object.keys(this.rooms);
		for (var k = 0; k < keys.length; k++) arr.push(this.rooms[keys[k]]);

		return arr;
	},

	addRoom: function (room, users) {
		if (this.rooms[room] !== undefined) return;

		if (users === null || users === undefined) users = [];

		this.rooms[room] = {
			name: room,
			users: users,
			messages: []
		};
	},

	removeRoom: function (room) {
		this.rooms[room] = undefined;
	},

	addMessageToRoom: function (room, message) {
		this.rooms[room].messages.push(message);
	},

	addMessage: function (rawmessage) {
		rawmessage = rawmessage.trim();
		var pMessage = parseMessage(rawmessage);

		var room = "server";
		var output = rawmessage;
		if (pMessage.command == "PRIVMSG" && pMessage.args.length >= 2) {
			room = pMessage.args[0];
			output = pMessage.nick + ": " + pMessage.args[1];
		} else if (pMessage.command == "JOIN" && pMessage.args.length >= 1) {
			room = pMessage.args[0];
			this.addUser(room, pMessage.nick);
			output = ">>> " + pMessage.nick + " has joined the channel.";
		} else if (pMessage.command == "PART" && pMessage.args.length >= 1) {
			room = pMessage.args[0];
			this.removeUser(room, pMessage.nick);
			output = "<<< " + pMessage.nick + " has left the channel.";
		}

		if (this.rooms[room] === undefined) this.addRoom(room);

		this.addMessageToRoom(room, output);
	},

	addUser: function (room, user) {
		if (this.rooms[room] === undefined) return;
		this.rooms[room].users.push(user);
	},

	removeUser: function (room, user) {
		//TODO
	}

};

//Helper methods to parse irc messages.

//[0:fullstring, 1: prefix, 2: command, 3: destination, 4: message contents]
var ircRegex = /:(\S+) (\S+) (\S+) ([:print:]+)/;

//[9:fullstring, 1: nick, 2: user, 3: host]
var userRegex = /(\S+)!~(\S+)@(\S+)/;

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
	retval.command = s.substring(0, end);

	//Parse the parameters by whhite space, everything after the ':' treated as one argument
	s = s.substring(end + 1, s.length);
	for (; s.length > 0;) {
		console.log("Test...");
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

	console.log("Parsed message: ", retval);
	return retval;
}

function parsePrefix(prefix) {
	var prefixarray = prefix.match(userRegex);

	console.log("parsePrefix: ", prefix, " parsed :", prefixarray);

	if (prefixarray != null && prefixarray.length > 3) return {
		prefix: prefixarray[0],
		nick: prefixarray[1],
		user: prefixarray[2],
		host: prefixarray[3]
	};
	return null;
}

module.exports = IRCStore;

},{}],3:[function(require,module,exports){
(function (global){
(function () {
	'use strict';
	/* Tabs provides a simple way of providing tabbed components. 
 
   General Features:
 	Tab names are taken from the children of <Tabs>. 
 		The default is to look for a property named 'tabName' on each child of <Tabs>. 
 		This can be modified by changing the property 'propName'.
 
 	By default, tabNames are used as keys and must be unique.
 		Set the property 'useKeys' to true to allow for non-unique tabNames.
 		If true, a property named 'key' that must be provided and be unique.
 	
     Multiple tabs being active is supported.
 		Set the property 'allowMultiple' to true to enable. 
 
   Props:
 	allowMultiple: bool - If true, multiple tabs can be selected. 
 	
 	propName: string, default: 'tabName' - The propname that will be used
 		to get the title of the tab. This must be unique among all children
 		unless useKeys is enabled.
 	
 	useKeys: bool - If true, Tabs requires children to have a prop value 'key' 
 		that are unique (but tabnames no longer need to be unique)		
 	
 	active: string, or array of strings - The current active tab by title 
 		(or key if useKeys=true). If allowMultiple=true, active will be an 
 		array of titles/keys	
 	
 	useState: bool, default: false - The user is normally responsible for maintaining 
 		the prop 'active' based on the value supplied to the onChange callback.
 		If useState is set to true, this component maintains the stateful value of the active tab.
 		The application does not need to update the property 'active'. This should only
 		be used if you have no interest in maintaining which tab is currently active.
 	
 	onChange: callback - The callback is called when the user clicks on a new tab. It
 		provides the tabName or key of the tab clicks. If allowMultiple is true,
 		onCHange will provide an array of the tabNames/keys that should be active.
 
 
   Example:
 		<Tabs useState={true}>
 			<Elem tabName='Action for Tab1!' />
 			<SomeOtherElem tabName='Tile for Tab#2' />
 			<div tabName='Tab Number 3!'>Interesting lack of content.</div>
 		</Tabs>
 */

	//Root code stolen from underscore.js (https://github.com/jashkenas/underscore/blob/master/underscore.js):
	// Establish the root object, `window` (`self`) in the browser, `global`
	// on the server, or `this` in some virtual machines. We use `self`
	// instead of `window` for `WebWorker` support.

	var root = typeof self == 'object' && self.self === self && self || typeof global == 'object' && global.global === global && global || this;

	var has_require = typeof require !== 'undefined';

	var React = root.React;
	if (typeof React === 'undefined') {
		if (has_require) React = require('react');else throw new Error('Tabs requires the React object be defined.');
	}

	var Tabs = React.createClass({
		displayName: 'Tabs',

		getDefaultProps: function () {
			return {
				active: null,
				allowMultiple: false,
				propName: 'tabName',
				onChange: null,
				useKeys: false,
				useState: false
			};
		},

		/* The initial state is copied over from the specified props.
     If allowMultiple=true, active becomes an array. 
  */
		getInitialState: function () {
			var p = this.props;
			var active = p.active;
			//Set active to an array if allowMultiple is true.
			if (p.allowMultiple) {
				if (active === null) active = [];else if (!Array.isArray(active)) active = [active];
			}
			return { active: active };
		},

		//When receiving new props, if tabs is not managing own state, update state.
		componentWillReceiveProps: function (newProps) {
			if (this.props.useState) return; //Let state manage things, not props.

			var active = newProps.active;
			if (newProps.allowMultiple) {
				if (active === null) active = [];else if (!Array.isArray(active)) active = [active];
			}
			this.setState({ active: active });
		},

		/* render() iterates through all the children, reading their tabname, 
     and saving active children to an array for display. 
   */render: function () {
			var self = this;
			var children = this.props.children;
			var activeChildren = [];

			return React.createElement(
				'div',
				null,
				React.createElement(
					'ul',
					{ className: 'tabs' },
					React.Children.map(children, function (child) {
						if (child === null) return;

						var name = self._getTabName(child);
						var key = self._getKey(child);
						var active = self._isActive(child);
						if (active) activeChildren.push(child);

						return React.createElement(
							'li',
							{ className: active ? 'active' : '',
								onClick: self._onClick.bind(null, key) },
							name
						);
					})
				),
				React.createElement(
					'div',
					{ className: 'tabcontent' },
					activeChildren
				)
			);
		},

		//Returns the name of the specified child as defined by propName.
		_getTabName: function (child) {
			return child.props[this.props.propName];
		},

		//Returns the key of the child (either the tabName, or the child
		//key if useKeys is turned on.
		_getKey: function (child) {
			if (this.props.useKeys) return child.key;else return this._getTabName(child);
		},

		//Returns true if this tab is listed as active in state.active.
		_isActive: function (child) {
			var key = this._getKey(child);
			if (this.props.allowMultiple) {
				return this.state.active.indexOf(key) >= 0;
			}

			return this.state.active === key;
		},

		//Calls the specified callbacks with the requested new state.
		_notifyCallbacks: function (newState) {
			var callbacks = this.props.onChange;

			if (callbacks !== null) {
				if (Array.isArray(callbacks)) callbacks.map(function (cb) {
					cb(newState);
				});else callbacks(newState);
			}
		},

		//Triggered when clicking a new tab. Triggers the onChange callback
		//if useState is true. the state is immediately updated to show the new tab
		//TODO: Reverse order - notify callbacks, allow them to prevent state change.
		_onClick: function (key) {
			var p = this.props;
			var active = this.state.active;
			var retval;
			if (p.allowMultiple) {
				var index = active.indexOf(key);
				var newVal = active.slice();
				if (index >= 0) newVal.splice(index, 1);else newVal.push(key);

				retval = { active: newVal };
				if (p.useState) this.setState(retval);
				this._notifyCallbacks(retval);
			} else if (active !== key) {
				retval = { active: key };
				if (p.useState) this.setState(retval);
				this._notifyCallbacks(retval);
			}
		}
	});

	//Export code stolen from underscore.js (https://github.com/jashkenas/underscore/blob/master/underscore.js):
	// Export the Tabs object for **Node.js**, with
	// backwards-compatibility for their old module API. If we're in
	// the browser, add `Tabs` as a global object.
	// (`nodeType` is checked to ensure that `module`
	// and `exports` are not HTML elements.)
	if (typeof exports != 'undefined' && !exports.nodeType) {
		if (typeof module != 'undefined' && !module.nodeType && module.exports) {
			exports = module.exports = Tabs;
		}
		exports.Tabs = Tabs;
	} else {
		root.Tabs = Tabs;
	}
})();
}).call(this,typeof global !== "undefined" ? global : typeof self !== "undefined" ? self : typeof window !== "undefined" ? window : {})
},{"react":"react"}]},{},[1]);
