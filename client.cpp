#include <iostream>
#include "QUIC.h"
#include <thread>
#include <chrono>

using namespace std;

void Receive(const char *str){
	cout << str << endl;
}

int main(int argc, char **argv) {
	QUIC* q = QUIC::getInstance();
	int id = q->connect("127.0.0.1", 8081, Receive);
	q->sendMsg(id, "hi world");
	this_thread::sleep_for(chrono::seconds(10));
	q->disconnect(id);
	return 0;
}
