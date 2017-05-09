package main

import (
	"fmt"
	"bufio"
	"os"
	"net"
)

func main() {
	fmt.Println("Welcome to my FTP client.")
	reader := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter Server Address")
	fmt.Print("-> ")
	reader.Scan()
	address := reader.Text()
	fmt.Println("Enter Server Port")
	fmt.Print("-> ")
	reader.Scan()
	port := reader.Text()
	fmt.Print("You entered " + address + ":" + port)
	conn, _ := net.Dial("tcp", address + ":" + port)
	message, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Print("Received from server: " + message)
}
