package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/getlantern/systray"
	"github.com/gorilla/websocket"
)

type WebSocketMessage struct {
	Type      string        `json:"type"`
	ProductID string        `json:"product_id"`
	Price     string        `json:"price"`
	Side      string        `json:"side"`
	Time      string        `json:"time"`
	TradeID   int           `json:"trade_id"`
	Size      string        `json:"size"`
	Message   string        `json:"message"`
	Channels  []interface{} `json:"channels"`
	Reason    string        `json:"reason"`
}

type BitcoinData struct {
	currentPrice float64
	holdings     float64
}

var holdingsFile string

func init() {
	holdingsFile = getHoldingsFilePath()
}

func getHoldingsFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Could not get home directory: %v", err)
		return "holdings.txt" // Fallback to current directory
	}

	// Create Application Support directory if it doesn't exist
	appSupportDir := filepath.Join(homeDir, "Library", "Application Support", "gobtc")
	if err := os.MkdirAll(appSupportDir, 0755); err != nil {
		log.Printf("Could not create app directory: %v", err)
		return "holdings.txt" // Fallback to current directory
	}

	return filepath.Join(appSupportDir, "holdings.txt")
}

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	data := &BitcoinData{holdings: loadHoldings()}

	systray.SetIcon(getIcon())
	systray.SetTitle("BTC")
	systray.SetTooltip("Bitcoin Price Tracker")

	// Add menu items
	mHoldings := systray.AddMenuItem("Set Holdings", "Enter your Bitcoin amount")
	mCurrentHoldings := systray.AddMenuItem("Current Holdings: ₿0.00000000", "Your Bitcoin amount")
	mCurrentHoldings.Disable() // Make it non-clickable
	systray.AddSeparator()
	mWorth := systray.AddMenuItem("Current Worth: $0.00", "Value of your holdings")
	mWorth.Disable() // Make it non-clickable
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the app")

	// Initialize the holdings display with loaded value
	if data.holdings > 0 {
		mCurrentHoldings.SetTitle(fmt.Sprintf("Current Holdings: ₿%.8f", data.holdings))
	}

	// Start WebSocket connection with shared data
	go connectWebSocket(data, mWorth)

	// Handle menu events
	go func() {
		for {
			select {
			case <-mHoldings.ClickedCh:
				cmd := exec.Command("osascript", "-e", `tell application "System Events" to display dialog "Enter your Bitcoin holdings:" default answer ""`)
				output, err := cmd.Output()
				if err != nil {
					log.Printf("Dialog error: %v", err)
					continue
				}

				parts := strings.Split(string(output), "text returned:")
				if len(parts) > 1 {
					input := strings.TrimSpace(strings.TrimRight(parts[1], "\n"))
					if holdings, err := strconv.ParseFloat(input, 64); err == nil {
						data.holdings = holdings
						// Save the new holdings
						if err := saveHoldings(holdings); err != nil {
							log.Printf("Error saving holdings: %v", err)
						}

						// Update displays
						mCurrentHoldings.SetTitle(fmt.Sprintf("Current Holdings: ₿%.8f", data.holdings))
						if data.currentPrice > 0 {
							worth := data.holdings * data.currentPrice
							mWorth.SetTitle(fmt.Sprintf("Current Worth: $%s", addThousandSeparators(worth)))
						}
					}
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func connectWebSocket(data *BitcoinData, mWorth *systray.MenuItem) {
	ws, _, err := websocket.DefaultDialer.Dial("wss://ws-feed.exchange.coinbase.com", nil)
	if err != nil {
		log.Fatal("WebSocket connection error:", err)
	}
	defer ws.Close()

	log.Println("Connected to Coinbase WebSocket")

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
		var msg WebSocketMessage
		if err := ws.ReadJSON(&msg); err != nil {
			log.Printf("WebSocket error: %v\n", err)
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

				data.currentPrice = floatPrice
				formattedPrice := addThousandSeparators(floatPrice)
				systray.SetTitle(fmt.Sprintf("₿ $%s", formattedPrice))

				// Update worth if holdings are set
				if data.holdings > 0 {
					worth := data.holdings * floatPrice
					mWorth.SetTitle(fmt.Sprintf("Current Worth: $%s", addThousandSeparators(worth)))
				}
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

// Add these functions to handle file operations
func saveHoldings(amount float64) error {
	return os.WriteFile(holdingsFile, []byte(fmt.Sprintf("%.8f", amount)), 0644)
}

func loadHoldings() float64 {
	data, err := os.ReadFile(holdingsFile)
	if err != nil {
		return 0.0 // Default to 0 if file doesn't exist
	}

	holdings, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
	if err != nil {
		return 0.0
	}
	return holdings
}
