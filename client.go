package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

type Server struct {
	host, port, message, r_code, data_host, data_port string
	ctrl_conn, data_conn                              net.Conn
	data_socket                                       net.Listener
	passive                                           bool //passive is true if the user sets up data port with PASV or EPSV
}

func (s *Server) address() string {
	return s.host + ":" + s.port
}

func (s *Server) connect() error {
	s.ctrl_conn, _ = net.Dial("tcp", s.address())
	s.getResponse()
	switch s.r_code {
	case "421":
		return FTPError{level: "fatal", what: s.message}
	case "220":
		return nil
	}
	return FTPError{level: "fatal", what: s.message}
}

func (s *Server) getResponse() {
	buf := make([]byte, 1024)
	n, _ := s.ctrl_conn.Read(buf)
	s.message = string(buf[:n-2])
	_code_and_message := strings.Split(s.message, " ")
	s.r_code = _code_and_message[0]
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

func (s *Server) setDataConnPRTE() {
	if s.passive {
		var _host string
		if s.data_host == "" || s.data_host == "127.0.0.1" || s.data_host == s.host {
			_host = ""
		} else {
			_host = s.data_host
		}
		s.data_conn, _ = net.Dial("tcp", _host+":"+s.data_port)
	} else {
		s.data_socket, _ = net.Listen("tcp", ":"+s.data_port)
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
	srvr.send("QUIT")
	os.Exit(0)
}

func PWD(args string) {
	srvr.send("PWD")
}

func LIST(args string) {
	var buf bytes.Buffer
	srvr.send("LIST " + args)
	if !srvr.passive {
		srvr.data_conn, _ = srvr.data_socket.Accept()
	}
	defer srvr.data_conn.Close()
	fmt.Println(srvr.message)
	io.Copy(&buf, srvr.data_conn)
	_bytes_read := len(buf.String())
	fmt.Println(buf.String()[:_bytes_read-2])
	srvr.getResponse()
}

func HELP(args string) {
	srvr.send("HELP")
}

func CDUP(args string) {
	srvr.send("CDUP")
}

func CWD(args string) {
	srvr.send("CWD " + args)
}

func NOOP(args string) {
	srvr.send("NOOP")
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
	file.Write(buf[:n-2])
	srvr.getResponse()
}

func PORT(args string) {
	srvr.passive = false
	_host_upper_lower := strings.Split(args, ",")
	host, upper, lower := _host_upper_lower[:4], _host_upper_lower[4], _host_upper_lower[5]
	srvr.setDataConnPRT(strings.Join(host, ","), upper, lower)
	srvr.send("PORT " + args)
	return
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

func EPSV(args string) {
	srvr.passive = true
	srvr.send("EPSV")
	_upper := strings.Index(srvr.message, "(|") + 2
	_lower := strings.Index(srvr.message, "|)")
	_address := srvr.message[_upper:_lower]
	_prtcl_host_port := strings.Split(_address, "|")
	srvr.data_host = _prtcl_host_port[1]
	srvr.data_port = _prtcl_host_port[2]
	srvr.setDataConnPRTE()
}

func EPRT(args string) {
	srvr.passive = false
	srvr.send("EPRT " + args)
	_prtcl_host_port := strings.Split(args, "|")
	srvr.data_host = _prtcl_host_port[2]
	srvr.data_port = _prtcl_host_port[3]
	fmt.Println(srvr.data_host)
	fmt.Println(srvr.data_port)
	srvr.setDataConnPRTE()
}

type FTPError struct {
	what, level string
}

func (e FTPError) Error() string {
	return fmt.Sprintf("%v: %v", e.level, e.what)
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
		"EPSV": EPSV,
		"EPRT": EPRT,
		"CDUP": CDUP,
		"CWD":  CWD,
		"HELP": HELP,
		"NOOP": NOOP,
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
	if err := srvr.connect(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(srvr.message)

	for {
		fmt.Print("command-with-args> ")
		reader.Scan()
		processCommand(reader.Text())
		fmt.Println(srvr.message)
	}
}
