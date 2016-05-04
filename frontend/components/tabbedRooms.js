var Tabs = require('tabs.react');
var Room = require('./room');

var IRC = require('../irc');

var TabbedRooms = React.createClass({
	propTypes: {
		rooms: React.PropTypes.arrayOf(React.PropTypes.instanceOf(IRC.Room)),
		activeTab: React.PropTypes.string.isRequired,
		onChange: React.PropTypes.func.isRequired
	},

	render: function(){
		var self = this;
		return (
			<Tabs active={this.props.activeTab} propName={'name'} onChange={this.props.onChange}>
				{this.props.rooms.map(function(room){
					return <Room name={room.Name()} users={room.Users()} messages={room.Messages()} key={room.Name()} />
				})}
			</Tabs>		
		)
	}
}); 

module.exports = TabbedRooms; 