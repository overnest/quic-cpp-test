CXX=g++
LDFLAGS=-g -Wall -std=c++11 -pthread

all: server client

server: QUIC.h
	$(CXX) -o server server.cpp QUIC.cpp ./quic_lib.so $(LDFLAGS)

client: QUIC.h
	$(CXX) -o client client.cpp QUIC.cpp ./quic_lib.so $(LDFLAGS)

clean:
	$(RM) server client
