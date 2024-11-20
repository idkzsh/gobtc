package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/getlantern/systray"
)

type CoinbaseResponse struct {
	Data struct {
		Amount string `json:"amount"`
	} `json:"data"`
}

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(getIcon())
	systray.SetTitle("BTC")
	systray.SetTooltip("Bitcoin Price Tracker")

	mRefresh := systray.AddMenuItem("Refresh", "Refresh Bitcoin price")
	mQuit := systray.AddMenuItem("Quit", "Quit the app")

	go func() {
		for {
			updatePrice()
			time.Sleep(10 * time.Second)
		}
	}()

	// Handle menu events
	go func() {
		for {
			select {
			case <-mRefresh.ClickedCh:
				updatePrice()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
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

func updatePrice() {
	price, err := getBitcoinPrice()
	if err != nil {
		systray.SetTitle("BTC Error")
		return
	}

	floatPrice, err := strconv.ParseFloat(price, 64)
	if err != nil {
		fmt.Println("Error converting price to float:", err)
		return
	}

	formattedPrice := addThousandSeparators(floatPrice)
	systray.SetTitle(fmt.Sprintf("₿ $%s", formattedPrice))
}

func getBitcoinPrice() (string, error) {
	resp, err := http.Get("https://api.coinbase.com/v2/prices/BTC-CAD/spot")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result CoinbaseResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.Data.Amount, nil
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
