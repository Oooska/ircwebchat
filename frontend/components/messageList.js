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

module.exports = MessageList;