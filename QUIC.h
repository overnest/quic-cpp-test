#pragma once

#include "libquic.h"

class QUIC
{
    private:
        /* Here will be the instance stored. */
        static QUIC* instance;
	static bool running;
	typedef void (*rptr)(const char *);
	static rptr callbacks[10000000];

        /* Private constructor to prevent instancing. */
        QUIC();
	/* Private method to convert char * to GoString */
	GoString toGoString(const char *str);
	/* Private method for threads to callback */
	void callBack(int id, const char *str);
	/* Method to run in thread to listen for incoming msgs */
	void connListen(int id);

    public:
        /* Static access method. */
        static QUIC* getInstance();
	/* Destructor */
	~QUIC();
	/* Public method to start listening for connections on a port */
	bool start(int port, rptr f(int));
	/* Shutdown all connections and stop listening for connections */
	void stop();
	/* Method to connect to an IP on a port */
	int connect(const char *ip, int port, void f(const char *));
	/* Method to check if connection is open */
	bool connStatus(int id);
	/* Method to get the IP address of a connection */
	char* getAddr(int id);
	/* Method to close a connection */
	void disconnect(int id);
	/* Method to send a message */
	bool sendMsg(int id, const char *message);
};
