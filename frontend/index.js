//index.js
var React = require('react');
var ReactDOM = require('react-dom')
var Tabs = require('tabs.react');
var IRCStore = require('./ircstore')

var IRCWebChat = React.createClass({
	getInitialState: function(){
		return {
			rooms: [{name: "server", users: ["server"], messages: ["Loading..."]}],

			activeTab: "server",
			input: {value: ""}
		}
	},

	componentWillMount: function(){
		IRCStore.addChangeListener(this.addMessage);
		IRCStore.start(window.location.host+"/chat/socket");
	},


	addMessage: function(newRooms){
		this.setState({rooms:  newRooms});
	},

	inputChange: function(event){
		this.setState({input: {value : event.target.value} });
	},

	sendMessage: function(event){
		event.preventDefault();

		var val = this.state.input.value;
		if(val.length > 0 && val[0] == '/')
			val = val.substring(1, val.length);
		else if(this.activeTab != "")
			val = "PRIVMSG "+this.state.activeTab+" :" + val;
		
		console.log("Sending message. Input: ", this.state.input.value, " Parsed to :",val)

		IRCStore.sendMessage(val);
		this.setState({input: {value : ''}});
	},

	_tabChanged: function(newValue){
		console.log("new tab: ", newValue)
		this.setState({activeTab: newValue.active});
	},

	render: function(){
		return (
			<div className="container-fluid">
				<TabbedRooms rooms={this.state.rooms} activeTab={this.state.activeTab} onChange={this._tabChanged} />
				<Input value={this.state.input.value} onChange={this.inputChange} onSend={this.sendMessage} />
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
			<ul className="nicklist col-xs-2">
				{rows}
			</ul>
		)
	}
});

var MessageList = React.createClass({
	propTypes: {
		messages: React.PropTypes.arrayOf(React.PropTypes.string).isRequired
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