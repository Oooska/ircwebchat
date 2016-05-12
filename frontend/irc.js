//Helper methods to parse irc messages. 
//[9:fullstring, 1: nick, 2: user, 3: host]
var userRegex = /(\S+)!(\S+)@(\S+)/

//Message represents an IRC message. The constructor parses the message
//and provides accessor methods to the various fields of a message
class Message{
    constructor(msg){
        this.message = msg;
        var rval = parseMessage(msg);
        this.prefix = rval.prefix;
        this.nick = rval.nick;
        this.user = rval.user;
        this.host = rval.host;
        this.command = rval.command.toUpperCase();
        this.args = rval.args;
        
        this._setDisplayText();
    }

    Prefix() {
        return this.prefix;
    }
    
    Nick() {
        return this.nick;
    }
    
    User() {
        return this.user;
    }
    
    Host() {
        return this.host;
    }
    
    Command() {
        return this.command;
    }
    
    Args(index) {
        if(index === undefined){ 
            return this.args;
        }
        return this.args[index];
    }
    
    toString() {
        return this.message;
    }
    
    DisplayText(){
        return this.displayText;
    }
    
    _setDisplayText(){
        var cmd = this.Command()
        if(cmd === "PRIVMSG"){
            var msg = this.Args(1);
            if(msg.startsWith(" ACTION")){
                this.displayText = msg.substring(7);
            } else {
                this.displayText = ": "+msg;
            }
        } else if(cmd === "JOIN"){
            this.displayText = " joined the room.";
        } else if(cmd === "PART"){
            this.displayText = " has left the room.";
        } else if(cmd === "QUIT"){
            this.displayText = " has quit: " + this.Args(0);
        } else {
            this.displayText = this.toString();
        }
    }
}

//Room represents an IRC channel and message queue.
class Room{
    constructor(name){
        this.name = name;
        this.users = {};
        this._usersArr = [];
        this.messages = [];
    }
    
    AddMessage(msg){
        this.messages.push(msg);
    }
    
    Name(){
        return this.name;
    }
    
    Messages(){ 
        return this.messages;
    }
    
    Users(){
        return this._usersArr;
    }
    
    AddUser(...users){
        //TODO: Add users more efficiently
        for(var i in users){
            var user = users[i];
            var name = user;
            if(user[0] == "@" || user[0] == "+"){
                name = user.substring(1);
            }
            console.log("Adding user: ", user)
            
            this.users[name] = user;
        }
        this._updateUserArr();
        
    }
    
    RemoveUser(...users){
        for(var i in users){
            var user = users[i];
            var name = user;
            if(user[0] == "@" || user[0] == "+"){
                name = user.substring(1);
            }
            delete this.users[name];
        }
        this._updateUserArr();
    }
    
    ClearUsers(){
        this.users = {};
        this._usersArr = [];
    }
    
    _updateUserArr(){
        this._usersArr = []
        for(var key in this.users){
            this._usersArr.push(this.users[key]);
        }
    }
}


//Helper function to parse a raw irc line
function parseMessage(message){
	var retval = {
		prefix: null, //nick!~user@host
		nick: null,
		user: null,
		host: null,
		command: null, //PRIVMSG or other command
		args: [] //Argument for command
	};

	//Parse the prefix if it is present (:Oooska1!~Oooska@knds.xdsl.dyn.ottcommunications.com)
	var s = message;
	if(s[0] == ":"){
		var end = s.indexOf(' ');
		retval.prefix = s.substring(1, s.indexOf(' '));

		var prefixArr = retval.prefix.match(userRegex);

		if(prefixArr != null && prefixArr.length >= 4){
			retval.nick = prefixArr[1];
			retval.user = prefixArr[2];
			retval.host = prefixArr[3];
		}

		s = s.substring(end+1, s.length);
	}

    
	//Parse the command
	var end = s.indexOf(' ');
    if(end < 0) //No arg commands will have no space
        end = s.length;
	retval.command = s.substring(0, end).toUpperCase();

	//Parse the parameters by white space, everything after the ':' treated as one argument
	s = s.substring(end+1, s.length);
	for (;s.length > 0;){
		if(s[0] == ':'){
			retval.args.push(s.substring(1,s.length));
			break;
		}

		end = s.indexOf(' ');
		if(end < 0){
			if(s.length > 0)
				retval.args.push(s);
			break;
		} else {
			retval.args.push(s.substring(0, end));
			if(end+1 >= s.length)
				break;
			s = s.substring(end+1, s.length);
		}
	}

	return retval;
}

//Helper function to parse the prefix of a raw irc line
function parsePrefix(prefix){
	var prefixarray = prefix.match(userRegex);

	if(prefixarray != null && prefixarray.length > 3)
	return {
		prefix: prefixarray[0],
		nick: prefixarray[1],
		user: prefixarray[2],
		host: prefixarray[3]
	};
	return null;
}

var IRC = {
    Room: Room,
    Message: Message
}

module.exports = IRC