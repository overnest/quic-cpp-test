#include <iostream>
#include "quic_lib.h"

using namespace std;

int main(int argc, char **argv) {
	GoString host = {"127.0.0.1", 9};
	startClient(host, 8081);
}
