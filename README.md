This is a jam session using websockets to increase throughput for enterprise workflows.

You are going to need the latest version of golang and docker on your machine. 

Before you get rolling you might want to do some reading on websockets.

https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API

I'd also check out the golang library I am using for managing the websockets.

https://github.com/gorilla/websocket

Before you run the project I always save myself some hastle by making sure docker is running on my computer.

`sudo docker ps -a`

You can start docker if it is not running.

`sudo systemctl start docker`

You can then start to have some fun with websockets.

cd in to the go-socket directory and run ./run.sh command in your terminal

The POC is behind localhost:8080/load-test

Hitting this endpoint in your browser or with a curl command will send 1000 concurrent calls to localhost:8081/process. 
You will receive a response once all 1000 requests have been persisted with their status.
You will not have to wait around for your requests to be processed which means you will receive confirmation all 1000 requests way quicker than you would expect.

The use case is asking for confirmation that a workflow has begun with the promise that my request will be persisted. Maybe the workflow will take milliseconds, maybe it will take a year. The promise is that my request has been received and that I have an id that I can track the request as well
as the abillity to check in on the status of my request. This makes for a lightweight api facing the client even when the processing in the background is expensive.

Where do websockets come in to play in this architecture? 

What if the processing of our worker takes minutes to hours due to external api calls that are having to retry after many failures? 

You'll notice the 2 second delay that is sumulated in the handler.go file in the worker on line 41.

Websockets give the ability of an infite open line of two way communication. If there are network issues, the request is persisted and can be attended either in an automated or manual fashion.

