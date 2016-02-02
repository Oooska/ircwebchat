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
		IRCStore.init(window.location.host+"/chat/socket");
	},


	addMessage: function(newRooms){
		this.setState({rooms:  newRooms});
	},

	inputChange: function(event){
		this.setState({input: {value : event.target.value} });
	},

	sendMessage: function(event){
		event.preventDefault();
		if(event.CharCode == 13)
			console.log("!!!ENTER PRESSED!!!")
		console.log("Received sendMessage event: ", event);
		IRCStore.sendMessage(this.state.input.value+"\r\n");
		this.setState({input: {value : ''}});
	},

	render: function(){
		return (
			<div className="container-fluid">
				<TabbedRooms rooms={this.state.rooms} />
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
			messages: React.PropTypes.arrayOf(React.PropTypes.string).isRequired
		}))
	},

	render: function(){
		return (
			<Tabs useState={true} propName={'name'} childClassNames="row col-xs-12">
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
			<div className="row" >
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
		//onChange: React.propTypes.Function
		//onSend:
		//value: React.PropTypes.string
	}, 
	render: function(){
		return (
			<div className="row">
				<input type="text" value={this.props.value} className="col-xs-11" onChange={this.props.onChange}  />
				<button className="col-xs-1" onClick={this.props.onSend}>Send</button>
			</div>
		)
	}
});

ReactDOM.render(
  <IRCWebChat></IRCWebChat>,
  document.getElementById('ircwebchat')
);