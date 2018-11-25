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
	jsonResponses := make(chan Response)
	go ping(ws, done)
	go returnWebsocketData(ws, jsonResponses, done)
	catchWebsocketData(ws, jsonResponses)

}

// poorly named but this is for lisening to imput from the site.
func catchWebsocketData(ws *websocket.Conn, jsonResponses chan<- Response) {
	defer ws.Close()
	ws.SetReadLimit(maxMessageSize)
	// Sets the timeout funtion
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}
		// all errors should be handeled in the func
		resp := parseJson(message)
		jsonResponses <- resp
		break
	}
}
func returnWebsocketData(ws *websocket.Conn, jsonResponses <-chan Response, done chan struct{}) {
	// sorry to whoever has to edit this
	// The general idea here is to have a goroutine waiting to write data to the  client
	// so whenever the the server reads data it pipes it into a chan which this method
	// reads and returns via ws.writeJson.
	for {
		select {
		case <-done:
			return
		case <-jsonResponses:
			err := ws.WriteJSON(jsonResponses)
			if err != nil {
				handle(err)
			}
		}
	}
}
func parseJson(message []byte) Response {
	// creates a map of strings to items, with strings as keys
	var r map[string]interface{}
	err := json.Unmarshal(message, &r)
	if err != nil {
		log.Fatal(err)
	}
	switch action := r["action"]; action {
	case "list":
		listOfPorts, err := listPorts()
		if err != nil {
			handle(err)
		}
		resp := Response{true, nil, listOfPorts}
		return resp

	case "connect":
		var p string = r["port"]
		err := connectPort(p)
		if err != nil {
			handle(err)
		}
		resp := Response{true, nil, nil}
		return resp
	case "write":
		var data []byte = r["data"]
		err := writePort(data)
		if err != nil {
			handle(err)
		}
		resp := Response{true, nil, nil}
		return resp

	case "disconnect":
		err := closePort()
		if err != nil {
			handle(err)
		}
		resp := Response{true, nil, nil}
		return resp
	default:
		// This should return if json is not properly formated.
		log.Fatal("invalid json")
	}
	return Response{}
}

type Request struct {
	action string `json:"action"`
	port   string `json:"port"`
	data   []byte `json:"data"`
}

type Response struct {
	success     bool
	err         error
	listOfPorts []string
}

func handle(err error) {
	// This will eventually handle internal errors
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
