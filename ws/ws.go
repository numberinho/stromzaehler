package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"zaehler/database"
	"zaehler/tracker"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID   string
	Conn *websocket.Conn
	Pool *Pool
}

type Message struct {
	Live   string `json:"Live"`
	Bezug  string `json:"Bezug"`
	Abgabe string `json:"Abgabe"`
}

type Pool struct {
	Register     chan *Client
	Unregister   chan *Client
	Clients      map[*Client]bool
	Zaehlerstand chan tracker.Zaehlerstand
}

func NewPool(wsChannel chan tracker.Zaehlerstand) *Pool {
	return &Pool{
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		Clients:      make(map[*Client]bool),
		Zaehlerstand: wsChannel,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (client *Client) HealthCheck() {
	defer func() {
		client.Pool.Unregister <- client
		client.Conn.Close()
	}()

	for {
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func (pool *Pool) Start() {
	for {
		select {
		//connect
		case client := <-pool.Register:
			pool.Clients[client] = true
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			fmt.Println("Clearing Channel: ", len(pool.Clients))
			break

		//disconnect
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			break

		//broadcast
		case zaehlerstand := <-pool.Zaehlerstand:

			for client := range pool.Clients {
				if err := client.Conn.WriteJSON(Message{
					Live:   fmt.Sprintf("%f", zaehlerstand.GetLive()),
					Bezug:  fmt.Sprintf("%f", zaehlerstand.Current.Bezug),
					Abgabe: fmt.Sprintf("%f", zaehlerstand.Current.Abgabe),
				}); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}

func RunWebserver(wsChannel chan tracker.Zaehlerstand, database *database.Database) {
	pool := NewPool(wsChannel)
	go pool.Start()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Fprintf(w, "%+v\n", err)
		}

		client := &Client{
			Conn: conn,
			Pool: pool,
		}

		pool.Register <- client
		client.HealthCheck()

	})

	http.HandleFunc("/history", getLastX(database))

	http.ListenAndServe(":8080", nil)
}

func getLastX(db *database.Database) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		x, _ := db.FetchLastNDailyData(6)
		log.Println(time.Now().Sub(start).String())

		w.Header().Set("Content-Type", "application/json")
		jData, err := json.Marshal(x)
		if err != nil {
			// handle error
		}
		w.Write(jData)
	}
}
