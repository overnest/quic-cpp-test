CXX=g++
LDFLAGS=-std=c++11 -pthread
LDLIBS=./quic_lib.so

server: QUIC.h
	$(CXX) -o server server.cpp QUIC.cpp ./quic_lib.so $(LDFLAGS)

client: QUIC.h
	$(CXX) -o client client.cpp QUIC.cpp ./quic_lib.so $(LDFLAGS)

clean:
	rm server client
