var MessageList = require('./messageList')
var NickList = require('./nickList')

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

module.exports = Room;