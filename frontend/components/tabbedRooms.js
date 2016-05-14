var Room = require('./room');

var IRC = require('../irc');

var TabbedRooms = React.createClass({
	propTypes: {
		rooms: React.PropTypes.arrayOf(React.PropTypes.instanceOf(IRC.Room)),
		activeRoom: React.PropTypes.instanceOf(IRC.Room),
		onChange: React.PropTypes.func.isRequired,
		onClose: React.PropTypes.func.isRequired,
		defaultName: React.PropTypes.string.isRequired
	},

	render: function(){
		var self = this;
		var activeRoom = this.props.activeRoom;
		var rooms = this.props.rooms;
		var onChange = this.props.onChange;
		var onClose = this.props.onClose;
		var defaultName = this.props.defaultName;
		return (
			<div>
				<ul className='tabs'>
					{rooms.map(function(room){
						return <li className={activeRoom.Name() === room.Name() ? "active" : "" } 
								   onClick={onChange.bind(null, room.Name())} 
								   key={room.Name()}>
							   		{room.Name()} 
									{room.Name() !== defaultName ? 
										<span onClick={onClose.bind(null, room.Name())}>[X]</span> : null }
							   </li>
					})}
				</ul>
				<Room name={activeRoom.Name()} users={activeRoom.Users()} messages={activeRoom.Messages()} />
			</div>
		)
	}
}); 

module.exports = TabbedRooms; 