package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"strconv"
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"encoding/binary"
	"fmt"
	"unsafe"
	"math/big"
	mathrand "math/rand"
	"time"
	"net"
	"sync"

	quicconn "github.com/marten-seemann/quic-conn"
)

var ln net.Listener
var conns sync.Map

func main(){
}

//export quic_startServer
func quic_startServer(port int) bool {
	tlsConf, err := generateTLSConfig()
	if err != nil {
		return false
	}

	ln, err = quicconn.Listen("udp",":" + strconv.Itoa(port), tlsConf)
	if err != nil {
		return false
	}

	return true
}

//export quic_listen
func quic_listen() int {
	conn, err := ln.Accept()
	if err != nil {
		return -1
	}

	mathrand.Seed(time.Now().UnixNano())
	id := mathrand.Intn(10000000)
	for quic_connExists(id) {
		id = mathrand.Intn(10000000)
	}

	conns.Store(id, conn)
	return id
}

//export quic_closeAll
func quic_closeAll(){
	ln.Close()
	conns.Range(func(key interface{}, value interface{}) bool {
		quic_close(key.(int))
		return true
	})
}

//export quic_startConn
func quic_startConn(ip string, port int) int {
	tlsConf := &tls.Config{InsecureSkipVerify: true}
	conn, err := quicconn.Dial(ip +":" + strconv.Itoa(port),tlsConf)
	if err != nil{
		return -1
	}

	mathrand.Seed(time.Now().UnixNano())
	id := mathrand.Intn(10000000)
	for quic_connExists(id) {
		id = mathrand.Intn(10000000)
	}

	conns.Store(id, conn)
	return id
}

//export quic_connExists
func quic_connExists(id int) bool {
	_, ok := conns.Load(id)
	return ok
}

//export quic_addr
func quic_addr(id int) *C.char {
	conn, ok := conns.Load(id)
	if ! ok {
		return nil
	}
	return C.CString(conn.(net.Conn).RemoteAddr().(*net.UDPAddr).IP.String())
}

//export quic_close
func quic_close(id int){
	conn, ok := conns.Load(id)
	if ok {
		conn.(net.Conn).Close()
		conns.Delete(id)
	}
}

func CArrayToByteSlice(array unsafe.Pointer, size int) []byte {
	var arrayptr = uintptr(array)
	var byteSlice = make([]byte, size)

	for i := 0; i < len(byteSlice); i++ {
		byteSlice[i] = byte(*(*C.char)(unsafe.Pointer(arrayptr)))
		arrayptr++
	}

	return byteSlice
}

//export quic_send
func quic_send(id int, message unsafe.Pointer, size int) bool {
	conn, ok := conns.Load(id)
	if !ok {
		return false
	}

	messageBytes := CArrayToByteSlice(message, size)
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(size))
	conn.(net.Conn).Write(bs)
	//for i := 0; i < len(messageBytes); i++ {
	//	conn.(net.Conn).Write([]byte{messageBytes[i]});
	//}
	conn.(net.Conn).Write(messageBytes)
	return true
}

func ByteSliceToCArray(byteSlice []byte) unsafe.Pointer {
	var array = unsafe.Pointer(C.calloc(C.size_t(len(byteSlice)),1))
	var arrayptr = uintptr(array)

	for i := 0; i < len(byteSlice); i++ {
		*(*C.char)(unsafe.Pointer(arrayptr)) = C.char(byteSlice[i])
		arrayptr++
	}

	return array
}

//export quic_receive
func quic_receive(id int) unsafe.Pointer {
	conn, ok := conns.Load(id)
	if !ok {
		return nil
	}

	reader := bufio.NewReader(conn.(net.Conn))
	bs := make([]byte, 4)
	for i := 0; i < 4; i++ {
		bs[i], _ = reader.ReadByte()
	}
	length := int(binary.LittleEndian.Uint32(bs))
	readBytes := make([]byte,length + 4)
	readBytes[0] = bs[0];
	readBytes[1] = bs[1];
	readBytes[2] = bs[2];
	readBytes[3] = bs[3];
	for i := 4; i < length + 4; i++ {
		readBytes[i], _ = reader.ReadByte()
	}

	return ByteSliceToCArray(readBytes)
}

func generateTLSConfig() (*tls.Config, error){
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil{
		return nil, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore: time.Now(),
		NotAfter: time.Now().Add(time.Hour),
		KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	b := pem.Block{Type: "CERTIFICATE", Bytes: certDER}
	certPEM := pem.EncodeToMemory(&b)

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}, nil
}
