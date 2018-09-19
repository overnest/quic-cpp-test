This repo will eventually be updated to follow the design patterns of Stellar/src/overlay networking files.
For now it contains a simple example where the client pings the server and it pings the same message back.
C++ wrapper files just call exported methods and feeds the correct C++ type defs as arguments.

Look inside the references folder to find the GoLang file that was used to generate quic_lib.so and quic_lib.h.
The folder also includes instructions to generate the files yourself.

Compile instructions:
1. $make server and $ make client
2. Run the binaries with ./server and ./client
