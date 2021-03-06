
ircwebchat is a webserver that hosts a web-based irc client. 
Users can login from multiple locations and access their IRC session.

The server is written in Go. 

The ircwebchat package exposes a Register() function that can be
used to register its http handlers, so it may be easily adapted
for use in other Go servers.

The frontend is written in Javascript. It currently requires websockets,
for communication. The UI is built using react.js


#Build Instructions
Building the server requires Go be installed and configured correctly.

##Server Instructions
To download:
git clone github.com/oooska/ircwebchat/ 
- or - 
go get github.com/oooska/ircwebchat/

###Run/Build
cd to $GOPATH/src/github.com/oooska/ircwebchat/server/
go run main.go

To build a binary, use go build or go install.


##Frontend Instructions
The frontend is build using react, and requires node.js to be installed.
A compiled copy of the front-end is included in the repo 
in ircwebchat/server/static/index.js

###Installation
Requires browserify and (optionally) watchify to be installed globally.
npm install -g browserify
npm install -g watchify

To install dependencies listed in frontend/package.json, cd into the 
frontend directory and run 'npm install'. 

###Build
npm run build - Compile the javascript one time and place it 
				at ../server/static/index.js 

npm run watch - Watches for changes in javascript files, re-compiles,
				and updates ../server/static/index.js

#Contributors:
	infina		- Thanks for the CSS updates, and being the first contributor.
	ubuntuguru	- For the idea