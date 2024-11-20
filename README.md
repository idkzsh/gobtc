# gobtc

A simple Bitcoin price display for the macOS menu bar.

## Build Instructions

1. Clone the repository
2. Build the binary:
   ```bash
   go build -o gobtc main.go
   ```
3. Create the app bundle:
   ```bash
   mkdir -p Bitcoin.app/Contents/MacOS
   cp gobtc Bitcoin.app/Contents/MacOS/
   cp Info.plist Bitcoin.app/Contents/
   chmod +x Bitcoin.app/Contents/MacOS/gobtc
   xattr -cr Bitcoin.app
   ```
