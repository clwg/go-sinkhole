package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

type ConnectionInfo struct {
	Timestamp       string `json:"timestamp"`
	SourceIP        string `json:"source_ip"`
	SourcePort      int    `json:"source_port"`
	DestinationPort int    `json:"destination_port"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: sinkhole <port1> <port2> <port3-range5> ...")
		os.Exit(1)
	}

	ports := parsePorts(os.Args[1:])

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			fmt.Println("\nClosing sinkhole servers.")
			os.Exit(0)
		}
	}()

	for _, port := range ports {
		go startSinkholeServer(port)
	}

	select {} // Block main goroutine to prevent exit
}

func parsePorts(args []string) []string {
	var ports []string

	for _, arg := range args {
		if strings.Contains(arg, "-") {
			r := strings.Split(arg, "-")
			start, _ := strconv.Atoi(r[0])
			end, _ := strconv.Atoi(r[1])

			for i := start; i <= end; i++ {
				ports = append(ports, strconv.Itoa(i))
			}
		} else {
			ports = append(ports, arg)
		}
	}

	return ports
}

func startSinkholeServer(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error listening on port", port, ":", err.Error())
		return
	}
	defer listener.Close()

	fmt.Println("Starting sinkhole server on port", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection on port", port, ":", err.Error())
			continue
		}

		go handleConnection(conn, port)
	}
}

func handleConnection(conn net.Conn, destPort string) {
	srcAddr := conn.RemoteAddr().(*net.TCPAddr)
	destPortInt, _ := strconv.Atoi(destPort)

	connectionInfo := ConnectionInfo{
		Timestamp:       time.Now().Format(time.RFC3339),
		SourceIP:        srcAddr.IP.String(),
		SourcePort:      srcAddr.Port,
		DestinationPort: destPortInt,
	}

	jsonData, err := json.Marshal(connectionInfo)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err.Error())
		return
	}

	fmt.Println(string(jsonData))
	conn.Close()
}
