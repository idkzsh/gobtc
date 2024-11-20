package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/getlantern/systray"
	"github.com/gorilla/websocket"
)

type WebSocketMessage struct {
	Type      string   `json:"type"`
	ProductID string   `json:"product_id"`
	Price     string   `json:"price"`
	Side      string   `json:"side"`
	Time      string   `json:"time"`
	TradeID   int      `json:"trade_id"`
	Size      string   `json:"size"`
	Message   string   `json:"message"`
	Channels  []string `json:"channels"`
	Reason    string   `json:"reason"`
}

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(getIcon())
	systray.SetTitle("BTC")
	systray.SetTooltip("Bitcoin Price Tracker")

	mQuit := systray.AddMenuItem("Quit", "Quit the app")

	// Start WebSocket connection
	go connectWebSocket()

	// Handle menu events
	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func connectWebSocket() {
	ws, _, err := websocket.DefaultDialer.Dial("wss://ws-feed.exchange.coinbase.com", nil)
	if err != nil {
		log.Fatal("WebSocket connection error:", err)
	}
	defer ws.Close()

	fmt.Println("Connected to Coinbase WebSocket")

	// Updated subscription format
	subscribe := map[string]interface{}{
		"type": "subscribe",
		"channels": []interface{}{
			map[string]interface{}{
				"name":        "ticker",
				"product_ids": []string{"BTC-USD"}, // We can try "BTC-CAD" here
			},
		},
	}

	if err := ws.WriteJSON(subscribe); err != nil {
		log.Fatal("Subscribe error:", err)
	}

	// Listen for messages
	for {
		// First, read the raw message to debug
		_, rawMsg, err := ws.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			continue
		}
		fmt.Printf("Raw message: %s\n", string(rawMsg))

		var msg WebSocketMessage
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			log.Println("JSON parse error:", err)
			continue
		}

		// Handle different message types
		switch msg.Type {
		case "error":
			log.Printf("WebSocket error: %s - %s\n", msg.Message, msg.Reason)
		case "subscriptions":
			log.Printf("Subscribed to channels: %v\n", msg.Channels)
		case "ticker":
			if msg.Price != "" {
				floatPrice, err := strconv.ParseFloat(msg.Price, 64)
				if err != nil {
					log.Println("Error converting price:", err)
					continue
				}

				formattedPrice := addThousandSeparators(floatPrice)
				systray.SetTitle(fmt.Sprintf("₿ $%s", formattedPrice))
			}
		}
	}
}

func addThousandSeparators(n float64) string {
	// Convert to string with 2 decimal places
	str := fmt.Sprintf("%.2f", n)

	// Split decimal and integer parts
	parts := strings.Split(str, ".")
	integer := parts[0]
	decimal := parts[1]

	// Add commas to integer part
	for i := len(integer) - 3; i > 0; i -= 3 {
		integer = integer[:i] + "," + integer[i:]
	}

	// Combine parts
	return integer + "." + decimal
}

func onExit() {
	// Clean up here
}

// getIcon returns a simple icon (you should replace this with your own icon)
func getIcon() []byte {
	// This is a minimal 16x16 pixel icon
	return []byte{
		0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF,
	}
}
