//index.js
var React = require('react');
var ReactDOM = require('react-dom');

var IRCStore = require('./ircstore');
var TabbedRooms = require('./components/tabbedRooms');
var Input = require('./components/textInput');

//The react interface for the IRC client.
var IRCWebChat = React.createClass({
	getInitialState: function(){
		return {
			rooms: IRCStore.Rooms(),
			activeRoom: IRCStore.Room(IRCStore.DefaultChannel),
			input: {value:""}	
		}
	},

	//Start the connection when the client mounts.
	componentWillMount: function(){
		IRCStore.AddChangeListener(this._updateIRCState);
		IRCStore.Start(window.location.host+"/chat/socket");
	},

	_updateIRCState: function(){
		var rooms = IRCStore.Rooms();
		var activeRoom = IRCStore.ActiveRoom();
		this.setState({rooms: rooms, activeRoom: activeRoom});
	},

	//sendMessage is called when the user hits enter or click send.
	//It tells the IRCStore to send the message.
	sendMessage: function(event){
		event.preventDefault();

		var val = this.state.input.value;
		if(val.length > 0 && val[0] == '/')
			val = val.substring(1, val.length);
		else if(this.state.activeRoom !== undefined && this.state.activeRoom.Name() != IRCStore.DefaultChannel){
			val = "PRIVMSG "+this.state.activeRoom.Name()+" :" + val;
		}
		
		IRCStore.SendMessage(val);
		this.setState({input: {value : ''}});
	},

	//Listens for the user switching tabs
	_tabChanged: function(newValue){
		IRCStore.SetActiveRoom(newValue);
	},

	//Kustebs for the close tab button being hit
	_closeTab: function(roomName){
		if(roomName === IRCStore.DefaultChannel)
			return;
		IRCStore.CloseRoom(roomName);
	},

	//Listens for changes to the Input box
	_inputChange: function(event){
		this.setState({input: {value : event.target.value} });
	},

	render: function(){
		return (
			<div className="container-fluid">
				<TabbedRooms rooms={this.state.rooms} activeRoom={this.state.activeRoom} onChange={this._tabChanged} onClose={this._closeTab} defaultName={IRCStore.DefaultChannel} />
				<Input value={this.state.input.value} onChange={this._inputChange} onSend={this.sendMessage} />
			</div>
		)
	}
});

ReactDOM.render(
  <IRCWebChat></IRCWebChat>,
  document.getElementById('ircwebchat')
);