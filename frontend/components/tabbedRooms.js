var Tabs = require('tabs.react');
var Room = require('./room');

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

module.exports = TabbedRooms; 