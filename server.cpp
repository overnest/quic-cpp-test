#include <iostream>
#include "QUIC.h"
#include <thread>
#include <chrono>

using namespace std;

QUIC* q;
int myID;
//function that handles the response
void receive(const char *str){
	cout << str << endl;
	q->sendMsg(myID, str);
}
//function that returns function pointer
typedef void (*rptr)(const char *);
rptr getReceiveFunc(int id){
	myID = id;
	return receive;
}

int main(int argc, char ** argv) {
	//get singleton
	q = QUIC::getInstance();
	//start the server
	q->start(8081, getReceiveFunc);
	//keep open for 20 min
	this_thread::sleep_for(chrono::seconds(1200));
	//close all connections and stop listening for new ones
	q->stop();
	return 0;
}
