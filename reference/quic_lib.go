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
	"time"
	"net"
	
	quicconn "github.com/marten-seemann/quic-conn"
)

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
	conn, err := ln.Accept()
	if err != nil {
		panic(err)
	}
	fmt.Println("Established connection")

	for {
		message := receive(conn)
		if err != nil {
			panic(err)
		}
		fmt.Print("Message from client: ", string(message) + "\n")
		//echo back
		send(conn, message)
	}
}

//export startClient
func startClient(ip string, port int){
	tlsConf := &tls.Config{InsecureSkipVerify: true}
	conn, err := quicconn.Dial(ip +":" + strconv.Itoa(port),tlsConf)
	if err != nil{
		panic(err)
	}

	message := "Ping from client"
	send(conn, []byte(message))
	fmt.Printf("Sending message: %s\n", message)
	//listen for reply
	answer := receive(conn)
	if err != nil {
		panic(err)
	}
	fmt.Print("Message from server: " + string(answer) + "\n")
}

func send(conn net.Conn, message []byte){
	length := len(message)
	bs := []byte{byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(length)}
	conn.Write(bs)
	conn.Write(message)
}

func receive(conn net.Conn) ([]byte){
	reader := bufio.NewReader(conn)
	a, err := reader.ReadByte()
	b, err := reader.ReadByte()
	c, err := reader.ReadByte()
	d, err := reader.ReadByte()

	if err != nil {
		panic(err)
	}
	length := int(d) | int(c << 8) | int(b << 16) | int(a << 24)
	readBytes := make([]byte,length)
	for i := 0; i < length; i++ {
		readBytes[i], err = reader.ReadByte()
	}
	
	return readBytes
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
