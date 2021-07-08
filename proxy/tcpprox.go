package main

import (
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"

	"github.com/goblimey/go-tools/dailylogger"
	reportfeed "github.com/goblimey/go-tools/proxy/reportfeed"
	reporter "github.com/goblimey/go-tools/statusreporter"
)

// Terminology:
// This is a Man In The Middle (MITM) NTRIP proxy intended to go between:
//
//		 an NTRIP Client on the (probably) local machine and
//		 an NTRP  Server on   a (probably) remote machine.
//
// The program variables and functions are named accordingly.
//
// Logging can be verbose or quiet.  It's verbose by default.  It can be set
// initially by options and at runtime by sending HTTP requests:
//    /status/loglevel/0
//    /status/loglevel/1
//
// The /status/report request displays the timestamp and contents of the last
// input and output buffers.

var log *dailylogger.Writer

var reportFeed *reportfeed.ReportFeed

func init() {
	log = dailylogger.New("./log", "dailylog.", ".txt")
}

func main() {

	localPort := flag.Int("p", 0, "Local Port to listen on")
	localHost := flag.String("l", "", "Local address to listen on")
	remoteHostPtr := flag.String("r", "", "Remote Server address host:port")
	configPtr := flag.String("c", "", "Use a config file (set TLS ect) - Commandline params overwrite config file")
	tlsPtr := flag.Bool("s", false, "Create a TLS Proxy")
	certFilePtr := flag.String("cert", "", "Use a specific certificate file")

	controlHost := flag.String("ca", "localhost", "hostname to listen on for http logging requests")
	controlPort := flag.Int("cp", 8080, "port to listen on for http logging requests")

	verbose := false
	flag.BoolVar(&verbose, "v", true, "verbose logging (shorthand)")
	flag.BoolVar(&verbose, "verbose", true, "verbose logging")

	quiet := false
	flag.BoolVar(&quiet, "q", false, "quiet logging (shorthand)")
	flag.BoolVar(&quiet, "quiet", false, "quiet logging")

	flag.Parse()

	if verbose {
		log.EnableLogging()
	}

	if quiet {
		log.DisableLogging()
	}

	fmt.Fprintf(log, "setting up routes\n")

	setConfig(*configPtr, *localPort, *localHost, *remoteHostPtr, *certFilePtr)

	if config.Remotehost == "" {
		fmt.Fprintf(os.Stderr, "[x] Remote host required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Fprintf(log, "setting up the status reporter\n")

	reportFeed = reportfeed.MakeReportFeed(log)

	proxyReporter := reporter.MakeReporter(reportFeed, *controlHost, *controlPort)

	proxyReporter.SetUseTextTemplates(true)

	// Start the HTTP server for control requests.
	go proxyReporter.StartService()

	// Start the main server for NTRIP traffic.
	startClientListener(*tlsPtr)
}

// connect to local first
//
func startClientListener(isTLS bool) {

	client := connectToClient(isTLS)
	defer func() { client.Close() }()

	fmt.Fprintf(log, "[*] Listening for Client call ...\n")

	for {
		call, err := client.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to accept call from client: %s\n", err)
			break
		}
		id := ids
		ids++
		fmt.Fprintf(log, "[*][%d]connection Accepted from: client %s\n", id, call.RemoteAddr())

		server := connectToServer(isTLS)
		fmt.Fprintf(log, "[*][%d] Connected to server: %s\n", id, server.RemoteAddr())

		go handleMessages(server, call, isTLS, id)
	}
}

func connectToClient(isTLS bool) (conn net.Listener) {
	var err error

	if isTLS == true {
		conn, err = tlsListen()
	} else {
		fmt.Fprintf(log, "listening on %s\n", fmt.Sprint(config.Localhost, ":", config.Localport))
		conn, err = net.Listen("tcp", fmt.Sprint(config.Localhost, ":", config.Localport))
	}

	if err != nil {
		panic("failed to connect to client: " + err.Error())
	}

	return conn
}

func connectToServer(isTLS bool) (conn net.Conn) {
	var err error

	if isTLS == true {
		conf := tls.Config{InsecureSkipVerify: true}
		conn, err = tls.Dial("tcp", config.Remotehost, &conf)
	} else {
		conn, err = net.Dial("tcp", config.Remotehost)
	}

	if err != nil {
		panic("failed to connect to server: " + err.Error())
	}
	return conn
}

func handleMessages(server, client net.Conn, isTLS bool, id int) {

	// Next bit needs coordination?
	go handleServerMessages(server, client, id)
	handleClientMessages(server, client, id)
	server.Close()
	client.Close()
}

func handleClientMessages(server, client net.Conn, id int) {
	for {
		data := make([]byte, 2048)
		n, err := client.Read(data)
		if n > 0 {
			fmt.Fprintf(log, "From Client [%d]:\n%s\n", id, hex.Dump(data[:n]))
			//fmt.Fprintf("From Client:\n%s\n",hex.EncodeToString(data[:n]))
			// Hang onto the buffer for reporting until the next one arrives
			reportFeed.RecordClientBuffer(&data, uint64(id), n)
			server.Write(data[:n])
		}
		if err != nil && err == io.EOF { // INCONSISTENT?
			fmt.Println(err)
			return
		}
	}
}

func handleServerMessages(server, client net.Conn, id int) {
	for {
		data := make([]byte, 2048)
		n, err := server.Read(data)
		if n > 0 {
			fmt.Fprintf(log, "From Server [%d]:\n%s\n", id, hex.Dump(data[:n]))
			//fmt.Fprintf("From Server:\n%s\n",hex.EncodeToString(data[:n]))
			// Hang onto the buffer for reporting until the next one arrives
			reportFeed.RecordServerBuffer(&data, uint64(id), n)
			client.Write(data[:n])
		}
		if err != nil && err != io.EOF { // INCONSISTENT?
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			break
		}
	}
}

func setConfig(configFile string, localPort int, localHost, remoteHost string, certFile string) {
	if configFile != "" {
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[-] Not a valid config file: %s\n", err.Error())
			os.Exit(1)
		}
		err = json.Unmarshal(data, &config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[-] Not a valid config file: %s\n", err.Error())
			os.Exit(1)
		}
	} else {
		config = Config{TLS: &TLS{}}
	}

	if certFile != "" {
		config.CertFile = certFile
	}

	if localPort != 0 {
		config.Localport = localPort
	}
	if localHost != "" {
		config.Localhost = localHost
	}
	if remoteHost != "" {
		config.Remotehost = remoteHost
	}
}
