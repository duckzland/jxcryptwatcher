
---

# 🪙 JXCryptWatcher

A desktop GUI application for displaying real-time cryptocurrency price tickers using data retrieved from [CoinMarketCap](https://coinmarketcap.com/). Built with [Fyne](https://fyne.io/) — a cross-platform GUI toolkit for Go.

---

## 🚀 Features

- Live crypto price updates
- Configurable panels
- Auto-generated configuration files
- Lightweight and fast native desktop app

---

## 🛠️ Installation

### 1. Install Fyne dependencies

Follow the official Fyne setup guide: [https://docs.fyne.io/started](https://docs.fyne.io/started)

For Ubuntu:

```bash
sudo apt install golang gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev
```

### 2. Install Fyne CLI tools (optional)

If you want to package the app as an installable binary:

```bash
go install fyne.io/tools/cmd/fyne@latest
```

### 3. Update your `$PATH`

Make sure Go binaries are accessible:

```bash
# Add this to your ~/.bashrc
export PATH="$HOME/go/bin:$PATH"
```

Then refresh your shell:

```bash
source ~/.bashrc
```

### 4. Build the application

Run the provided build script:

```bash
./build.sh
```

This will:
- Download all required Go modules
- Compile the app
- Generate the `jxcryptwatcher` executable

---

## ⚙️ Configuration

The app requires three configuration files for normal operation:

- `config.json`
- `panels.json`
- `cryptos.json`

These files are automatically generated on first launch and saved to:

```
~/.config/jxcryptwatcher/config.json
~/.config/jxcryptwatcher/panels.json
~/.config/jxcryptwatcher/cryptos.json
```
> **Note:** The paths mentioned apply to **Ubuntu**. Other operating systems may use different locations for user-specific configuration files.  
> 
> - **Ubuntu/Linux**: `~/.config/jxcryptwatcher/`
> - **macOS**: `/Users/<username>/jxcryptwatcher/`
> - **Windows**: `C:\Users\<username>\jxcryptwatcher\`
> 
> Please adjust the paths accordingly based on your operating system and user environment.

### 📁 Example Configurations

You can find sample configuration files in the `examples/` directory:

```
examples/config_example.json
examples/panels_example.json
```

### 🔄 Refreshing Crypto Data

The `cryptos.json` file is auto-generated using data from CoinMarketCap.  
To refresh the list of available cryptocurrencies, simply delete the file:

```bash
rm ~/.config/jxcryptwatcher/cryptos.json
```

It will be re-created on the next app launch.

---

## 🧩 Notes

- Internet connection is required to fetch live data from CoinMarketCap.
- CoinMarketCap enforces rate limits on API requests. To avoid exceeding these limits, please configure a minimum delay of 60 seconds between each API call. Being considerate with request frequency helps ensure stable access and prevents temporary bans.
- This application has been tested primarily on **Ubuntu**. While it may work on other operating systems, additional installation steps or adjustments might be required depending on your platform.

---
