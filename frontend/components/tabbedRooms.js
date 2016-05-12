var Room = require('./room');

var IRC = require('../irc');

var TabbedRooms = React.createClass({
	propTypes: {
		rooms: React.PropTypes.arrayOf(React.PropTypes.instanceOf(IRC.Room)),
		activeRoom: React.PropTypes.instanceOf(IRC.Room),
		onChange: React.PropTypes.func.isRequired
	},

	render: function(){
		var self = this;
		var activeRoom = this.props.activeRoom;
		var rooms = this.props.rooms;
		var onChange = this.props.onChange;
		return (
			<div>
				<ul className='tabs'>
					{rooms.map(function(room){
						return <li className={activeRoom.Name() === room.Name() ? "active" : "" } 
								   onClick={onChange.bind(null, room.Name())} 
								   key={room.Name()}>
							   		{room.Name()}
							   </li>
					})}
				</ul>
				<Room name={activeRoom.Name()} users={activeRoom.Users()} messages={activeRoom.Messages()} />
			</div>
		)
	}
}); 

module.exports = TabbedRooms; 