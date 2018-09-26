#include <iostream>
#include <thread>
#include "quic_lib.h"

using namespace std;

GoString toGoString(const char *str){
	int len = 0;
	while(str[len] != '\0') len++;
	GoString retVal = {str, len};
	return retVal;
}

void listenConn(int id){
	while(true){
		char *received = receive(id);
		if(received == NULL){
			break;
		}
		cout << "Received from client: " << received << endl;
		send(id, toGoString(received));
	}
}

int main(int argc, char ** argv) {
	startServer(8081);
	while(true){
		int id = listen();
		thread new_thread (listenConn, id);
		new_thread.detach();
	}

	return 0;
}
