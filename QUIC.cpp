#include "QUIC.h"
#include "libquic.h"
#include <thread>
#include <iostream>

/* Null, because instance will be initialized on demand. */
QUIC* QUIC::instance = 0;
bool QUIC::running = false;
typedef void (*rptr)(const char *);
/* Empty, because function pointers will be stored when threads are created */
rptr QUIC::callbacks[10000000] = {};

QUIC::QUIC()
{}

GoString QUIC::toGoString(const char *str){
	int len = 0;
	while(str[len] != '\0') len++;
	return {str, len};
}

void QUIC::callBack(int id, const char *str)
{
	callbacks[id](str);
}

void QUIC::connListen(int id)
{
	while(true)
	{
		char *received = quic_receive(id);
		if(received == NULL)
			break;
		callBack(id, received);
	}
	callBack(id, NULL);
}

QUIC* QUIC::getInstance()
{
    if (instance == 0)
    {
        instance = new QUIC();
    }

    return instance;
}

QUIC::~QUIC()
{
	if(running)
		stop();
}

bool QUIC::start(int port, rptr f(int))
{
	//prevent more than one listener
	if(running)
	{
		return false;
	}
	running = quic_startServer(port);
	if(!running)
	{
		return false;
	}
	std::thread t([=] {
		while(true){
			int id = quic_listen();
			if(id < 0)
			{
				break;
			}
			callbacks[id] = f(id);
			std::thread new_thread([=] {connListen(id);});
			new_thread.detach();
		}
		running = false;
	});
	t.detach();
	return true;
}

void QUIC::stop()
{
	quic_closeAll();
}

int QUIC::connect(const char *ip, int port, void f(const char *))
{
	int id = quic_startConn(toGoString(ip), port);
	if(id >= 0)
	{
		callbacks[id] = f;
		std::thread new_thread([=] {connListen(id);});
		new_thread.detach();
	}
	return id;
}

bool connStatus(int id)
{
	return quic_connExists(id);
}

char* QUIC::getAddr(int id)
{
	return quic_addr(id);
}

void QUIC::disconnect(int id)
{
	quic_close(id);
}

bool QUIC::sendMsg(int id, const char *message)
{
	return quic_send(id, toGoString(message));
}
