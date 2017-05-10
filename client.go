package main

import (
	"fmt"
	"bufio"
	"os"
	"net"
	"strings"
	"strconv"
)

type Server struct {
	host, port, message, d_message, data_address string
	ctrl_conn, data_conn net.Conn
	data_socket net.Listener
}

func (s *Server) address() string  {
	return s.host + ":" + s.port
}

func (s *Server) connect() {
	s.ctrl_conn, _ = net.Dial("tcp", s.address())
	s.getResponse()
}

func (s *Server) getResponse() {
	s.message, _ = bufio.NewReader(s.ctrl_conn).ReadString('\n')
	//if s.data_conn != nil {
	//	s.d_message,_ = bufio.NewReader(s.data_conn).ReadString('\n')
	//	s.data_conn.Close()
	//	s.data_conn = nil
	//}
}

func (s *Server) send(msg string) {
	s.ctrl_conn.Write([]byte(msg + "\r\n"))
}

func (s *Server) setDataConnPRT(host, upper, lower string) {
	_upper, _ := strconv.Atoi(upper)
	_lower, _ := strconv.Atoi(lower)

	ctrl_port := (_upper << 8) | _lower
	strings.Replace(host, ",", ".",3)
	fmt.Println(host)
	fmt.Println(strconv.Itoa(ctrl_port))
	s.data_address = host + ":" + strconv.Itoa(ctrl_port)
	fmt.Println(s.data_address)
	fmt.Println("-----------------")
}

//func (s *Server) getDataResponse(){
//
//}

var (
	command2Func = map[string]func(Server, string){
		"USER": USER,
		"PASS": PASS,
		"QUIT": QUIT,
		"PWD" : PWD,
		"PORT": PORT,
		"LIST": LIST,
	}
	srvr Server
)

func main() {
	var host, port string;
	reader := bufio.NewScanner(os.Stdin)

	fmt.Println("Welcome to my FTP client.")
	if len(os.Args) != 3 {
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

	srvr = Server{host: host, port: port}
	fmt.Println("Server has address " + srvr.address())
	srvr.connect()
	fmt.Print("Server message " + srvr.message)

	for{
		fmt.Print("command-with-args> ")
		reader.Scan()
		processCommand(reader.Text())
		srvr.getResponse()
		fmt.Print("Response: " + srvr.message)
		if srvr.d_message != "" {
			fmt.Print(srvr.d_message)
			srvr.d_message = ""
		}
	}
}

func processCommand(cli string){
	var command, args string
	_split := strings.SplitN(cli, " ", 2)
	command = _split[0]
	if len(_split) > 1{
		args = _split[1]
	}
	command2Func[command](srvr, args)
}

func USER(s Server, args string){
	srvr.send("USER " + args)
}

func PASS(s Server, args string){
	srvr.send("PASS " + args)
}

func QUIT(s Server, args string){
	srvr.send("QUIT " + args)
}

func PWD(s Server, args string){
	srvr.send("PWD " + args)
}

func PORT(s Server, args string){
	_host_upper_lower := strings.Split(args, ",")
	host, upper, lower := _host_upper_lower[:4], _host_upper_lower[4], _host_upper_lower[5]
	srvr.setDataConnPRT(strings.Join(host, ","), upper, lower)
	srvr.send("PORT " + args)
	return
}

func LIST(s Server, args string){
	ln, _ := net.Listen("tcp", ":2020")
	srvr.send("LIST " + args)
	conn, _ := ln.Accept()
	fmt.Println(bufio.NewReader(conn).ReadString('\n'))
}
