This repo will eventually be updated to follow the design patterns of Stellar/src/overlay networking files.
For now it contains a simple example where the client pings the server and it pings the same message back.
C++ wrapper files just call exported methods and feeds the correct C++ type defs as arguments.

Look inside the references folder to find the GoLang file that was used to generate quic_lib.so and quic_lib.h.
The folder also includes instructions to generate the files yourself.

Compile instructions:
1. `$make server` and `$make client`
2. Run the binaries with `./server` and `./client`

Use:

To create a connection with a server use the startClient(ip string, port int) function.
To listen to connections on a port use the startServer(port int) function.

As of my last push sending, receiving, and listening will no longer be called from Go.
It will need to be called from C++ like in the examples.

The id used as an argument in the send, receive and close functions refers to the id of a connection when it is created.
listen() and startClient() create their a connection and return the connection id that can be used to call functions on it.

send(int id, GoString p0) is pretty straight forward, just convert to GoString in example given

close(int id) is pretty straight forward as well

receive(int id) needs to be run in a loop like this:
```C++
while(true){
	 char *received = receive(id);
	 if(received == NULL){
	   break;
	 }
	 //do whatever
}
```

listen() needs to be run like this
```C++
startServer(8081);
while(true){
	int id = listen();
	//do whatever
}
```
startServer(int port) must be called before entering the loop

receive() will return a null pointer if the connection is broken so you will need to check after each read to break the loop.
