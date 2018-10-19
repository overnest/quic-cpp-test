#include <iostream>
#include "QUIC.h"
#include <thread>
#include <chrono>

using namespace std;

class QUICPeer
{
public:
	QUIC* quic_ptr;
	int quic_id;
	QUICPeer(const char *addr, int port)
	{
		quic_ptr = QUIC::getInstance();
		auto callback = bind(&QUICPeer::recvMessage, this, placeholders::_1);
		quic_id = quic_ptr->connect(addr, port, callback);
	}

	void send(const char *str)
	{
		quic_ptr->sendMsg(quic_id, str);
	}

	void recvMessage(const char *str)
	{
		if(str != NULL){
			cout << str << endl;
		}
	}

	void drop()
	{
		quic_ptr->disconnect(quic_id);
	}
};

int main(int argc, char **argv) {
	auto qp = QUICPeer("127.0.0.1", 8081);
	qp.send("hi world");
	//keep open for 10s
	this_thread::sleep_for(chrono::seconds(10));
	qp.drop();
	return 0;
}
