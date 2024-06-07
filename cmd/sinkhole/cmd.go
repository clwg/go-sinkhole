package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	jsonlogger "github.com/clwg/go-rotating-logger"
)

type ConnectionInfo struct {
	Timestamp       string `json:"timestamp"`
	Protocol        string `json:"protocol"`
	SourceIP        string `json:"source_ip"`
	SourcePort      int    `json:"source_port"`
	DestinationPort int    `json:"destination_port"`
}

type AppConfig struct {
	LoggerConfig jsonlogger.LoggerConfig
	Protocol     string
	Ports        []string
}

func main() {
	appConfig := parseFlags()

	jsonLogger, err := jsonlogger.NewLogger(appConfig.LoggerConfig)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			fmt.Println("\nClosing sinkhole servers.")
			os.Exit(0)
		}
	}()

	for _, port := range appConfig.Ports {
		go startSinkholeServer(appConfig.Protocol, port, jsonLogger)
	}

	select {} // Block main goroutine to prevent exit
}

func parseFlags() AppConfig {
	var config AppConfig

	filenamePrefix := flag.String("filenamePrefix", "sinkholeserver", "Prefix for log filenames")
	logDir := flag.String("logDir", "./logs", "Directory for log files")
	maxLines := flag.Int("maxLines", 100000, "Maximum number of lines per log file")
	rotationTime := flag.Int("rotationTime", 60, "Log rotation time in minutes")
	protocol := flag.String("protocol", "", "Protocol to use (tcp or udp)")
	ports := flag.String("ports", "", "Comma-separated list of ports or port ranges (e.g., 8000,8001-8005)")
	flag.Parse()

	if *protocol != "tcp" && *protocol != "udp" {
		fmt.Println("Protocol must be either 'tcp' or 'udp'")
		os.Exit(1)
	}

	if *ports == "" {
		fmt.Println("Ports must be specified")
		os.Exit(1)
	}

	config.LoggerConfig = jsonlogger.LoggerConfig{
		FilenamePrefix: *filenamePrefix,
		LogDir:         *logDir,
		MaxLines:       *maxLines,
		RotationTime:   time.Duration(*rotationTime) * time.Minute,
	}
	config.Protocol = *protocol
	config.Ports = parsePorts(*ports)

	return config
}

func parsePorts(ports string) []string {
	var portList []string
	args := strings.Split(ports, ",")

	for _, arg := range args {
		if strings.Contains(arg, "-") {
			r := strings.Split(arg, "-")
			start, _ := strconv.Atoi(r[0])
			end, _ := strconv.Atoi(r[1])

			for i := start; i <= end; i++ {
				portList = append(portList, strconv.Itoa(i))
			}
		} else {
			portList = append(portList, arg)
		}
	}
	return portList
}

func startSinkholeServer(protocol, port string, jsonLogger *jsonlogger.Logger) {
	if protocol == "tcp" {
		startTCPSinkholeServer(port, jsonLogger)
	} else if protocol == "udp" {
		startUDPSinkholeServer(port, jsonLogger)
	} else {
		fmt.Println("Unknown protocol:", protocol)
	}
}

func startTCPSinkholeServer(port string, jsonLogger *jsonlogger.Logger) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error listening on port", port, ":", err.Error())
		return
	}
	defer listener.Close()

	fmt.Println("Starting TCP sinkhole server on port", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection on port", port, ":", err.Error())
			continue
		}

		go handleTCPConnection(conn, port, jsonLogger)
	}
}

func handleTCPConnection(conn net.Conn, destPort string, jsonLogger *jsonlogger.Logger) {
	srcAddr := conn.RemoteAddr().(*net.TCPAddr)
	destPortInt, _ := strconv.Atoi(destPort)

	connectionInfo := ConnectionInfo{
		Timestamp:       time.Now().Format(time.RFC3339),
		Protocol:        "tcp",
		SourceIP:        srcAddr.IP.String(),
		SourcePort:      srcAddr.Port,
		DestinationPort: destPortInt,
	}

	jsonLogger.Log(connectionInfo)

	conn.Close()
}

func startUDPSinkholeServer(port string, jsonLogger *jsonlogger.Logger) {
	addr := net.UDPAddr{
		Port: parseInt(port),
		IP:   net.ParseIP("0.0.0.0"),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println("Error listening on port", port, ":", err.Error())
		return
	}
	defer conn.Close()

	fmt.Println("Starting UDP sinkhole server on port", port)

	for {
		handleUDPConnection(conn, parseInt(port), jsonLogger)
	}
}

func handleUDPConnection(conn *net.UDPConn, destPort int, jsonLogger *jsonlogger.Logger) {
	buf := make([]byte, 1024)
	_, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println("Error reading from connection:", err.Error())
		return
	}

	connInfo := ConnectionInfo{
		Timestamp:       time.Now().Format(time.RFC3339),
		Protocol:        "udp",
		SourceIP:        addr.IP.String(),
		SourcePort:      addr.Port,
		DestinationPort: destPort,
	}

	jsonLogger.Log(connInfo)

	_, err = conn.WriteToUDP([]byte("true"), addr)
	if err != nil {
		fmt.Println("Error writing response:", err.Error())
		return
	}
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
