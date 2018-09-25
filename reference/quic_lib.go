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
	//"math/rand"
	"time"
	"net"
	
	quicconn "github.com/marten-seemann/quic-conn"
)

var conn net.Conn

//var ln net.Listener

//conns := make(map[string]net.Conn)

func main(){

	serverCmd := flag.Bool("s", false, "server")
	clientCmd := flag.Bool("c", false, "client")
	flag.Parse()

	if *serverCmd {
		startServer(8081)
	}
	if *clientCmd {
		startClient("127.0.0.1",8081)
	}
}

//export startServer
func startServer(port int){
	tlsConf, err := generateTLSConfig()
	if err != nil {
		panic(err)
	}

	ln, err := quicconn.Listen("udp",":" + strconv.Itoa(port), tlsConf)
	if err != nil {
		panic(err)
	}

	fmt.Println("Waiting for incoming connection")
	conn, err = ln.Accept()
	if err != nil {
		panic(err)
	}
	fmt.Println("Established connection")
}

//export listen
/*func listen() int {
	conn, err := ln.Accept()
	if err != nil {
		panic(err)
	}
	
	id := rand.Intn(10000000)
	while if inside already choose different one

	conns[id] = conn
	return id
}*/

//export startClient
func startClient(ip string, port int){
	tlsConf := &tls.Config{InsecureSkipVerify: true}
	var err error
	conn, err = quicconn.Dial(ip +":" + strconv.Itoa(port),tlsConf)
	if err != nil{
		panic(err)
	}
}

//export close
func close(){
	conn.Close()
}

//export send
func send(message string){
	messageBytes := []byte(message)
	length := len(messageBytes)
	bs := []byte{byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(length)}
	//conn := conns[id]
	conn.Write(bs)
	conn.Write(messageBytes)
}

//export receive
func receive() *C.char {
	reader := bufio.NewReader(conn)
	a, err := reader.ReadByte()
	b, err := reader.ReadByte()
	c, err := reader.ReadByte()
	d, err := reader.ReadByte()

	if err != nil {
		//panic(err)
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
