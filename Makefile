server:
	g++ -o server server.cpp ./quic_lib.so -std=c++11 -pthread

client:
	g++ -o client client.cpp ./quic_lib.so

clean:
	rm server client
