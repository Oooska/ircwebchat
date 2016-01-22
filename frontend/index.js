//index.js
//ircwebchat loads React and ReactDOM from a cdn, they will be available globally
//var React = require('react');
//var ReactDOM = require('react-dom');



var IRCWebChat = React.createClass({
	render: function(){
		return (
			<div>
				Hello World. Start of IRCWebChat Client...
			</div>
		)
	}
})






ReactDOM.render(
  <IRCWebChat />,
  document.getElementById('ircwebchat')
);


