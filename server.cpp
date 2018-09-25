#include <iostream>
#include "quic_lib.h"

using namespace std;

GoString toGoString(const char *str){
	int len = 0;
	while(str[len] != '\0') len++;
	GoString retVal = {str, len};
	return retVal;
}

int main(int argc, char ** argv) {
	startServer(8081);
	while(true){
		char *received = receive();
		if(received == NULL){
			break;
		}
		cout << "Received from client: " << received << endl;
		send(toGoString(received));
	}
}
