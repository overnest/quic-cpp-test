#pragma once

#include "libquic.h"

#include <functional>

class QUIC
{
    private:
        /* Here will be the instance stored. */
        static QUIC* instance;
	static bool running;
        /* Private constructor to prevent instancing. */
        QUIC();
	/* Private method to convert char * to GoString */
	GoString toGoString(const char *str);
	/* Private method for threads to callback */
	void callBack(int id, const char *str);

    public:
        /* Static access method. */
        static QUIC* getInstance();
	/* Destructor */
	~QUIC();
	/* Public method to start listening for connections on a port */
	bool start(int port, std::function<std::function<void(const char*)>(int)>);
	/* Shutdown all connections and stop listening for connections */
	void stop();
	/* Method to connect to an IP on a port */
	int connect(const char *ip, int port, std::function<void(const char*)> f);
	/* Method to check if connection is open */
	bool connStatus(int id);
	/* Method to get the IP address of a connection */
	char* getAddr(int id);
	/* Method to close a connection */
	void disconnect(int id);
	/* Method to send a message */
	bool sendMsg(int id, const char *message);
};
