package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Server struct {
	host, port, message, d_message, data_host, data_port string
	ctrl_conn, data_conn                                 net.Conn
	data_socket                                          net.Listener
	passive                                              bool
}

func (s *Server) address() string {
	return s.host + ":" + s.port
}

func (s *Server) connect() {
	s.ctrl_conn, _ = net.Dial("tcp", s.address())
	s.getResponse()
}

func (s *Server) getResponse() {
	buf := make([]byte, 1024)
	n, _ := s.ctrl_conn.Read(buf)
	s.message = string(buf[:n])
}

func (s *Server) send(msg string) {
	s.ctrl_conn.Write([]byte(msg + "\r\n"))
	s.getResponse()
}

func (s *Server) setDataConnPRT(host, upper, lower string) {
	_upper, _ := strconv.Atoi(upper)
	_lower, _ := strconv.Atoi(lower)

	s.data_port = strconv.Itoa((_upper << 8) | _lower)
	strings.Replace(host, ",", ".", 3)
	s.data_host = host
	if s.passive {
		s.data_conn, _ = net.Dial("tcp", ":"+s.data_port)
	} else {
		s.data_socket, _ = net.Listen("tcp", ":"+srvr.data_port)
	}
}

var (
	command2Func = map[string]func(string){
		"USER": USER,
		"PASS": PASS,
		"QUIT": QUIT,
		"PWD":  PWD,
		"PORT": PORT,
		"LIST": LIST,
		"PASV": PASV,
		"RETR": RETR,
		//"EPSV": EPSV,
		//"EPRT": EPRT,
		"CDUP": CDUP,
		"CWD":  CWD,
		"HELP": HELP,
		//"NOOP": NOOP,
	}
	srvr Server
)

func main() {
	var host, port string
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
	} else {
		host = os.Args[1]
		port = os.Args[2]
	}

	srvr = Server{host: host, port: port}
	fmt.Println("Server has address " + srvr.address())
	srvr.connect()
	fmt.Print("Server message " + srvr.message)

	for {
		fmt.Print("command-with-args> ")
		reader.Scan()
		processCommand(reader.Text())
		fmt.Print("Response: " + srvr.message)
	}
}

func processCommand(cli string) {
	var command, args string
	_split := strings.SplitN(cli, " ", 2)
	command = _split[0]
	if len(_split) > 1 {
		args = _split[1]
	}
	_func, ok := command2Func[command]
	if !ok {
		fmt.Println("Sorry command is not currently supported")
		return
	}

	_func(args)
}

func USER(args string) {
	srvr.send("USER " + args)
}

func PASS(args string) {
	srvr.send("PASS " + args)
}

func QUIT(args string) {
	srvr.send("QUIT " + args)
	os.Exit(0)
}

func PWD(args string) {
	srvr.send("PWD " + args)
}

func PORT(args string) {
	srvr.passive = false
	_host_upper_lower := strings.Split(args, ",")
	host, upper, lower := _host_upper_lower[:4], _host_upper_lower[4], _host_upper_lower[5]
	srvr.setDataConnPRT(strings.Join(host, ","), upper, lower)
	srvr.send("PORT " + args)
	return
}

func LIST(args string) {
	srvr.send("LIST " + args)
	if !srvr.passive {
		srvr.data_conn, _ = srvr.data_socket.Accept()
	}
	fmt.Print("Response: " + srvr.message)
	buf := make([]byte, 1024)
	n, _ := srvr.data_conn.Read(buf)
	fmt.Println(string(buf[:n]))
	srvr.getResponse()
}

func PASV(args string) {
	srvr.passive = true
	srvr.send("PASV " + args)
	_upper := strings.Index(srvr.message, "(") + 1
	_lower := strings.Index(srvr.message, ")")
	address := srvr.message[_upper:_lower]
	_address := strings.Split(address, ",")
	host, upper, lower := _address[0], _address[1], _address[2]
	srvr.setDataConnPRT(string(host), string(upper), string(lower))
}

func HELP(args string) {
	srvr.send("HELP " + args)
}

func CDUP(args string) {
	srvr.send("CDUP " + args)
}

func CWD(args string) {
	srvr.send("CWD " + args)
}

func RETR(args string) {
	srvr.send("RETR " + args)
	if !srvr.passive {
		srvr.data_conn, _ = srvr.data_socket.Accept()
	}
	fmt.Print("Response: " + srvr.message)
	_upper := strings.Index(srvr.message, "(") + 1
	_lower := strings.Index(srvr.message, " bytes)")
	_size, _ := strconv.Atoi(srvr.message[_upper:_lower])
	buf := make([]byte, _size)
	n, _ := srvr.data_conn.Read(buf)
	file, _ := os.Create(args)
	file.Write(buf[:n])
	srvr.getResponse()
}
