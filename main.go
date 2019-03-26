package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const MIN = 1
const MAX = 100

var clients []Client

type Client struct {
	Ip string
}

func random() int {
	return rand.Intn(MAX-MIN) + MIN
}

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", printClients).Methods("GET")

	return muxRouter
}

func printClients(w http.ResponseWriter, r *http.Request) {
	printSlice(clients)
	bytes, err := json.MarshalIndent(clients, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(bytes))
}

func run() error {
	mux := makeMuxRouter()
	httpAddr := "8012"
	log.Println("Listening on", httpAddr)

	s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func handleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		temp := strings.TrimSpace(string(netData))
		if temp == "STOP" {
			break
		}

		result := strconv.Itoa(random()) + "\n"
		c.Write([]byte(string(result)))
	}

	c.Close()
}

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide a port number!")
		return
	}

	PORT := ":" + arguments[1]
	l, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	rand.Seed(time.Now().Unix())
	first := Client{"127.0.0.1:8201"}
	clients = append(clients, first)

	go run()
	for {
		c, err := l.Accept()
		connection := Client{c.RemoteAddr().String()}
		clients = append(clients, connection)
		printSlice(clients)

		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c)
	}

}

func printSlice(s []Client) {
	fmt.Printf("len=%d cap=%d %v\n", len(s), cap(s), s)
}
