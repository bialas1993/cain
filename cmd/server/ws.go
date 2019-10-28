package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

type msg struct {
	Num int
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile("./web/index.html")
	if err != nil {
		fmt.Println("Could not open file.", err)
	}
	fmt.Fprintf(w, "%s", content)
}

func wsHandler(connections chan int, w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	go echo(conn, connections)
}

func echo(conn *websocket.Conn, clients chan int) {
	for {
		_, buf, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading msg.", err)
		}

		m := string(buf)
		fmt.Printf("Got message: %#v\n", m)

		newClientsQuantity, err := strconv.Atoi(m)
		if err != nil {
			fmt.Println("Can not convert massage")
			return
		}

		go func(c int) {
			clients <- c
		}(newClientsQuantity)

		if err = conn.WriteMessage(websocket.TextMessage, []byte(m)); err != nil {
			fmt.Println(err)
		}
	}
}
