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

module.exports = NickList;