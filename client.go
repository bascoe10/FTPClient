package main

import (
	"fmt"
	"bufio"
	"os"
	"net"
)

type Server struct {
	host, port, message string
	conn net.Conn
}

func (s *Server) address() string  {
	return s.host + ":" + s.port
}

func (s *Server) connect() {
	s.conn, _ = net.Dial("tcp", s.address())
	s.message, _ = bufio.NewReader(s.conn).ReadString('\n')
}


func main() {
	var host, port string;
	fmt.Println("Welcome to my FTP client.")
	if len(os.Args) != 3 {
		reader := bufio.NewScanner(os.Stdin)
		fmt.Println("Enter Server Address")
		fmt.Print("-> ")
		reader.Scan()
		host = reader.Text()
		fmt.Println("Enter Server Port")
		fmt.Print("-> ")
		reader.Scan()
		port = reader.Text()
	}else {
		host = os.Args[1]
		port = os.Args[2]
	}

	srvr := Server{host: host, port: port}
	fmt.Println("Server has address " + srvr.address())
	srvr.connect()
	fmt.Print("Server message " + srvr.message)
	//conn, _ := net.Dial("tcp", host + ":" + port)
	//message, _ := bufio.NewReader(conn).ReadString('\n')
	//fmt.Print("Received from server: " + message)
}
