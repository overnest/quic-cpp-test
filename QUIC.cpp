#include "QUIC.h"

#include <thread>
#include <iostream>

/* Null, because instance will be initialized on demand. */
QUIC* QUIC::instance = 0;
bool QUIC::running = false;

QUIC::QUIC()
{}

GoString QUIC::toGoString(const char *str){
	int len = 0;
	while(str[len] != '\0') len++;
	return {str, len};
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

bool QUIC::start(int port, std::function<std::function<void(const char*)>(int)> f)
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
			auto g = f(id);
			std::thread new_thread([=] {
				while(true)
				{
					char *received = quic_receive(id);
					if(received == NULL)
						break;
					g(received);
				}
				g(NULL);
			});
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

int QUIC::connect(const char *ip, int port, std::function<void(const char*)> f)
{
	int id = quic_startConn(toGoString(ip), port);
	if(id >= 0)
	{
		std::thread new_thread([=] {
			while(true)
			{
				char *received = quic_receive(id);
				if(received == NULL)
					break;
				f(received);
			}
			f(NULL);
		});
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
