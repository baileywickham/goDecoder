package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

func main() {
	serve()
}

//server functionality
const (
	// this format must have comments between items, all of them are costs
	maxMessageSize = 8192
	// timeout on pongs
	pongWait = 60 * time.Second
	// Period of pings sent
	pingPeriod = (pongWait * 9) / 10
	// time allowed to write
	wrieWait = 10 * time.Second
)

var upgrader = websocket.Upgrader{}

func serve() {
	//entry point for handling server
	http.HandleFunc("/", serveConnection)
	http.HandleFunc("/ws", serveWS)
	addr := "127.0.0.1:8080"
	log.Fatal(http.ListenAndServe(addr, nil))
}

func serveConnection(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}
	if r.Method != "GET" {
		http.Error(w, "Improper request", http.StatusNotFound)
	}
	http.ServeFile(w, r, "home.html")
}

func serveWS(w http.ResponseWriter, r *http.Request) {
	//upgrade to Websocket from normal http
	// Main function, w is unupgraded
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade", err)
	}
	defer ws.Close()
	// no idea what this does
	done := make(chan struct{})
	go ping(ws, done)
	//go returnWebsocketData(ws, jsonResponses, done)
	catchWebsocketData(ws)
}

// poorly named but this is for lisening to imput from the site.
func catchWebsocketData(ws *websocket.Conn) {
	defer ws.Close()
	// Checks if port is open, then closes
	defer closePort()
	ws.SetReadLimit(maxMessageSize)
	// Sets the timeout funtion
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}
		// all errors should be handeled in the func
		data := make(chan Response)
		go parseJSON(message, data)
		go returnData(ws, data)
	}
}

func returnData(ws *websocket.Conn, data <-chan Response) {
	for {
		select {
		case d := <-data:
			err := ws.WriteJSON(d)
			if err != nil {
				// Handles problems with returning to the websocket.
				// Cannot return, must log.
				handle(err)
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}
func parseJSON(message []byte, data chan<- Response) {
	// creates a map of strings to items, with strings as keys
	var r Request
	err := json.Unmarshal(message, &r)
	if err != nil {
		log.Fatal("unmarshal error", err)
	}
	switch action := r.Action; action {
	case "list":
		listOfPorts, err := listPorts()
		if err != nil {
			//handle(err)
			resp := Response{false, err.Error(), nil, nil}
			data <- resp
		} else {
			resp := Response{true, "", listOfPorts, nil}
			data <- resp
		}

	case "connect":
		// Close to force closed open connections
		closePort()
		var p = r.Port
		err := connectPort(p)
		if err != nil {
			//handle(err)
			resp := Response{false, err.Error(), nil, nil}
			data <- resp
		} else {
			resp := Response{true, "", nil, nil}
			data <- resp
			go readPort(data)
		}

	case "write":
		var d = r.Data
		err := writePort([]byte(d))
		if err != nil {
			handle(err)
		} else {

			resp := Response{true, "", nil, nil}
			data <- resp
		}

	case "disconnect":
		closePort()
		resp := Response{true, "", nil, nil}
		data <- resp
	default:
		// This should return if json is not properly formated.
		log.Fatal("invalid json", r)
	}
}

// Request from websocket
type Request struct {
	Action string `json:"action"`
	Port   string `json:"port"`
	Data   string `json:"data"`
}

// Repsponse to websocket
type Response struct {
	Success     bool
	Err         string
	ListOfPorts []string
	Data        []byte
}

func handle(err error) {
	// This will eventually handle internal errors
	closePort()
	panic(err)
}

// a keepalive function. This pings to make sure the browser is still active
func ping(ws *websocket.Conn, done chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(wrieWait)); err != nil {
				log.Println("ping error: ", err)
			}
		case <-done:
			return
		}
	}
}
