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
	QUICPeer(int ID)
	{
		quic_ptr = QUIC::getInstance();
		quic_id = ID;
	}

	void send(const char *str)
	{
		quic_ptr->sendMsg(quic_id, str);
	}

	void recvMessage(const char *str)
	{
		if(str != NULL){
			cout << str << endl;
			quic_ptr->sendMsg(quic_id, str);
		}
	}

	void drop()
	{
		quic_ptr->disconnect(quic_id);
	}
};

class PeerDoor
{
public:
	QUIC* quic_ptr;
	void listen(int port)
	{
		quic_ptr = QUIC::getInstance();
		auto callback = bind(&PeerDoor::quicKnock, this, placeholders::_1);
		quic_ptr->start(port, callback);
	}
	
	void stopListening()
	{
		quic_ptr->stop();
	}

	std::function<void(const char*)> quicKnock(int ID)
	{
		
		auto callback = bind(&QUICPeer::recvMessage,QUICPeer(ID), placeholders::_1);
		return callback;
	}
};

int main(int argc, char ** argv) {
	auto pd = PeerDoor();
	pd.listen(8081);
	//keep open for 20 min
	this_thread::sleep_for(chrono::seconds(1200));
	//close all connections and stop listening for new ones
	pd.stopListening();
	return 0;
}
