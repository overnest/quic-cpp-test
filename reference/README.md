Compile Instructions

This file relies on https://github.com/lucas-clemente/quic-go and https://github.com/marten-seemann/quic-conn

Make sure file path for both repos are located inside go/src/github.com and follow instructions to get dependencies for quic-go

I compiled quic_lib.go from go/src/github.com/marten-seemann/quic-conn/example and made changes to the given example as follows:

1. Changed from sending strings to []byte slices
2. Separated code into methods to be exported with cgo with proper flags

To compile yourself and update the shared object libraries do the following:

1. $cp quic_lib.go $HOME/go/src/github.com/marten-seeman/example
2. $cd $HOME/go/src/github.com/marten-seeman/example
3. $go build -o quic_lib.so -buildmode=c-shared quic_lib.go
4. Copy quic_lib.so and quic_lib.h into main directory of this repo
5. Use the provided make file $make server and $make client
