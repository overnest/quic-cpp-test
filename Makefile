CXX=g++
LDFLAGS=-g -Wall -std=c++14 -pthread
LDLIBS=-lquic

all: server client

server: QUIC.h
	$(CXX) -o server server.cpp QUIC.cpp $(LDLIBS) $(LDFLAGS)

client: QUIC.h
	$(CXX) -o client client.cpp QUIC.cpp $(LDLIBS) $(LDFLAGS)

clean:
	$(RM) server client
