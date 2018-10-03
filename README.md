This repo has been updated to use a Singleton design pattern that has two purposes:
1. To prevent duplicate GoLang containers.
2. To allow external calls to the library without dealing with Go types.

Since Go cannot export complex structs such as network connections and socket listeners to C++, I was forced to keep them inside the Go container like this:
```Go
var ln net.Listener
var conns sync.Map
```
By keeping these global they do not need to be included in function parameters and return types which allow them to exported with cgo. Since C++ can't see any of this happening calling the exported library must be done from a Singleton.

Look inside the references folder to find the GoLang file that was used to generate quic_lib.so and quic_lib.h.
The folder also includes instructions to generate the files yourself.

Example instructions:
1. `$make server` and `$make client` or just run `$make all`
2. Run the binaries with `./server` and `./client`

To include in another project you will need 4 files:
1. QUIC.cpp
2. QUIC.h
3. quic_lib.so
4. quic_lib.h
Modify your makefile and follow examples below.

Use:
To start listening for connections:
```C++
QUIC* q = QUIC::getInstance();
bool success = q->start(8081, getReceiveFunc);
```
The second parameter to start needs a function that returns a function pointer. It will also return true or false.
```C++
void receive(const char *str){
	if(str == NULL){
		//handle connection closed
	}
	cout << str << endl;
 	//do whatever
}
typedef void (*rptr)(const char *);
rptr getReceiveFunc(int id){
	//do whatever
	return receive;
}
```
Receiving function will get a null value when the connection is broken. The threads inside the Singleton are not tracked because they will auto terminate when they receive a null value (the disconnect signal). The connection at the GoLang level is also deleted automatically. However the stored function pointers will remain in the Singleton but will never be called since no thread for it exists anymore. If another connection reuses the same ID it will simply be overriden.
When you're done call the following:
```C++
q->stop();
```
This will stop the server and safely close all connections. If you forget to call this don't worry because it gets called by the singleton destructor anyway.

To start a connection:
```C++
int id = q->connect("127.0.0.1", 8081, receive);
```
Make sure to store the returned integer ID since you will need it to call all functions related to the connection. The second argument is similar that of starting the server but you just need to provide a function pointer directly this time.

To send:
```C++
bool success = q->sendMsg(id, "hi world");
```
And to close:
```C++
q->disconnect(id);
```
