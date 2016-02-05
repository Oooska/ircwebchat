//index.js
var React = require('react');
var ReactDOM = require('react-dom')
var Tabs = require('tabs.react');
var IRCStore = require('./ircstore')

//The react interface for the IRC client.
//TODO: Break into multiple component files.
var IRCWebChat = React.createClass({
	getInitialState: function(){
		return {
			rooms: [{name: "Server", users: [], messages: ["Loading..."]}],
			activeTab: "Server",
			input: {value: ""}
		}
	},

	//Start the connection when the client mounts.
	componentWillMount: function(){
		IRCStore.addChangeListener(this.addMessage);
		IRCStore.start(window.location.host+"/chat/socket");
	},


	//addMessage is called by the store when there's updated state to pass down.
	addMessage: function(newRooms){
		this.setState({rooms:  newRooms});
	},


	//sendMessage is called when the user hits enter or click send.
	//It tells the IRCStore to send the message.
	sendMessage: function(event){
		event.preventDefault();

		var val = this.state.input.value;
		if(val.length > 0 && val[0] == '/')
			val = val.substring(1, val.length);
		else if(this.state.activeTab != "" && this.state.activeTab != "Server"){
			console.log("this.activeTab: ", this.activeTab)
			val = "PRIVMSG "+this.state.activeTab+" :" + val;
		}
		
		console.log("Sending message. Input: ", this.state.input.value, " Parsed to :",val)

		IRCStore.sendMessage(val);
		this.setState({input: {value : ''}});
	},

	//Listens for the user switching tabs
	_tabChanged: function(newValue){
		console.log("new tab: ", newValue)
		this.setState({activeTab: newValue.active});
	},

	//Listens for changes to the Input box
	_inputChange: function(event){
		this.setState({input: {value : event.target.value} });
	},

	render: function(){
		return (
			<div className="container-fluid">
				<TabbedRooms rooms={this.state.rooms} activeTab={this.state.activeTab} onChange={this._tabChanged} />
				<Input value={this.state.input.value} onChange={this._inputChange} onSend={this.sendMessage} />
			</div>
		)
	}
});

var TabbedRooms = React.createClass({
	propTypes: {
		rooms: React.PropTypes.arrayOf(React.PropTypes.shape({
			name: React.PropTypes.string.isRequired,
			users: React.PropTypes.arrayOf(React.PropTypes.string).isRequired,
			messages: React.PropTypes.arrayOf(React.PropTypes.string).isRequired,
		})),

		activeTab: React.PropTypes.string.isRequired,
		onChange: React.PropTypes.func.isRequired
	},

	render: function(){
		var self = this;
		return (
			<Tabs active={this.props.activeTab} propName={'name'} onChange={this.props.onChange}>
				{this.props.rooms.map(function(room){
					return <Room name={room.name} users={room.users} messages={room.messages} key={room.name} />
				})}
			</Tabs>		
		)
	}
}) 

var Room = React.createClass({
	propTypes: {
			name: React.PropTypes.string.isRequired,
			users: React.PropTypes.arrayOf(React.PropTypes.string).isRequired,
			messages: React.PropTypes.arrayOf(React.PropTypes.string).isRequired
	},

	render: function(){
		var isRoom = this.props.name[0] == '#';
		return (
			<div className="row">
				<MessageList messages={this.props.messages} />
				{isRoom ? <NickList users={this.props.users} /> : null}
			</div>
		)
	}
})

var NickList = React.createClass({ 
	propTypes: {
		users: React.PropTypes.arrayOf(React.PropTypes.string).isRequired
	},

	render: function(){
		var rows = [];
		for(var k=0; k < this.props.users.length; k++)
			rows.push(<li className="nick" key={k}>{this.props.users[k]}</li>)
		return (
			<div className="nicklist">
				<ul className="col-xs-2">
					{rows}
				</ul>
			</div>
		)
	}
});

var MessageList = React.createClass({
	propTypes: {
		messages: React.PropTypes.arrayOf(React.PropTypes.string).isRequired
	},

	componentWillUpdate: function(){
		//Determine if we're at the bottom of the message list
		var node = ReactDOM.findDOMNode(this);
		this.atBottom = node.scrollTop + node.offsetHeight === node.scrollHeight;
	},

	componentDidUpdate: function(){
		if(this.atBottom){ 
			//If we're at the bottom, make sure we stay at the bottom
			var node = ReactDOM.findDOMNode(this);
			node.scrollTop = node.scrollHeight;
		}
	},

	render: function(){
		var rows = [];
		for(var k=0; k < this.props.messages.length; k++)
			rows.push( <span key={k}>{this.props.messages[k]}</span> )
		return (
			<div className="messagelist col-xs-10">
				{rows}
			</div>
		)
	}
});

var Input = React.createClass({
	propTypes: {
		onChange: React.PropTypes.func,
		onSend: React.PropTypes.func,
		value: React.PropTypes.string
	}, 
	render: function(){
		return (
			<div className="ircinput row">
				<input type="text" value={this.props.value} className="col-xs-11" 
				       onKeyDown={this.checkForSend} onChange={this.props.onChange}  />
				<button className="col-xs-1" onClick={this.props.onSend}>Send</button>
			</div>
		)
	},

	checkForSend: function(event){
		if(event.key == 'Enter')
			this.props.onSend(event);
	}
});

ReactDOM.render(
  <IRCWebChat></IRCWebChat>,
  document.getElementById('ircwebchat')
);