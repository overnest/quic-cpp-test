server:
	g++ -o server server.cpp ./quic_lib.so

client:
	g++ -o client client.cpp ./quic_lib.so

clean:
	rm server client
