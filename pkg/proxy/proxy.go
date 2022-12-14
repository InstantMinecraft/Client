package proxy

import (
	"encoding/hex"
	"fmt"
	"github.com/instantminecraft/client/pkg/mcserver"
	"github.com/instantminecraft/client/pkg/server"
	"io"
	"log"
	"net"
)

const (
	PORT = 25585
)

func Start() {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		log.Fatalln("Couldn't start tcp proxy at port", PORT, "=>", err)
	}
	defer l.Close()

	log.Println("Started proxy on port", PORT)
	log.Println("HTTP and Minecraft Clients should connect to this proxy")

	for {
		if conn, err := l.Accept(); err == nil {
			go acceptClient(conn)
		}
	}
}

func acceptClient(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 2)
	_, err := conn.Read(buf)
	if err != nil {
		return
	}

	signature := hex.EncodeToString(buf)

	// select target port
	var targetPort int
	if isMinecraftConnection(signature) {
		// Proxy connection to minecraft server like nothing happened
		targetPort = mcserver.PORT
	} else {
		// Proxy connection to local HTTP server
		targetPort = server.PORT
	}

	targetConnection, err := net.Dial("tcp", fmt.Sprintf(":%d", targetPort))
	if err != nil {
		// oh oh!
		return
	}

	// write the first read bytes
	if _, err = targetConnection.Write(buf); err != nil {
		return
	}

	// Start proxy
	go func() {
		if _, err := io.Copy(conn, targetConnection); err != nil {
			return
		}
	}()
	defer targetConnection.Close()
	if _, err := io.Copy(targetConnection, conn); err != nil {
		return
	}
}

func isMinecraftConnection(signature string) bool {
	// The first 4 digits area always "1000" if a minecraft client tries to connect
	return signature == "1000"
}
