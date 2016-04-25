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

module.exports = Input;