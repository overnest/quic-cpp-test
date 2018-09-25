#include <iostream>
#include "quic_lib.h"

using namespace std;

GoString toGoString(const char *str){
	int len = 0;
	while(str[len] != '\0') len++;
	GoString retVal = {str, len};
	return retVal;
}


int main(int argc, char **argv) {
	startClient(toGoString("127.0.0.1"), 8081);
	send(toGoString("hi world"));
	cout << receive() << endl;
	close();
}
