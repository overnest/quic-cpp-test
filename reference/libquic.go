package main

import "C"

import (
	"strconv"
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
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

	serverCmd := flag.Bool("s", false, "server")
	clientCmd := flag.Bool("c", false, "client")
	flag.Parse()

	if *serverCmd {
		quic_startServer(8081)
	}
	if *clientCmd {
		quic_startConn("127.0.0.1",8081)
	}
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

//export quic_close
func quic_close(id int){
	conn, ok := conns.Load(id)
	if ok {
		conn.(net.Conn).Close()
		conns.Delete(id)
	}
}

//export quic_send
func quic_send(id int, message string) bool {
	conn, ok := conns.Load(id)
	if !ok {
		return false
	}
	messageBytes := []byte(message)
	length := len(messageBytes)
	bs := []byte{byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(length)}
	conn.(net.Conn).Write(bs)
	conn.(net.Conn).Write(messageBytes)
	return true
}

//export quic_receive
func quic_receive(id int) *C.char {
	conn, ok := conns.Load(id)
	if !ok {
		return nil
	}

	reader := bufio.NewReader(conn.(net.Conn))
	a, err := reader.ReadByte()
	b, err := reader.ReadByte()
	c, err := reader.ReadByte()
	d, err := reader.ReadByte()

	if err != nil {
		conns.Delete(id)
		return nil
	}
	length := int(d) | int(c << 8) | int(b << 16) | int(a << 24)
	readBytes := make([]byte,length)
	for i := 0; i < length; i++ {
		readBytes[i], err = reader.ReadByte()
	}

	return C.CString(string(readBytes))
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
