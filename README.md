Showing GUI for displaying Crypto price ticker with data retrieved from coinmarketcap


#### Installation

Only tested under Ubuntu, get the required development packages

```
sudo apt install golang gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev
```

Then after cloning this repository, initialize go modules

```
cd jxcryptwatcher
go mod init main
go get fyne.io/fyne/v2@latest
go get fyne.io/fyne/v2
go get fyne.io/fyne/v2/app
go get fyne.io/fyne/v2/canvas
go get fyne.io/fyne/v2/container
go get fyne.io/fyne/v2/layout
go get fyne.io/fyne/v2/widget
go mod tidy
go install fyne.io/tools/cmd/fyne@latest
```

Might need to update $PATH to point to the fyne binary by editing ~/.bashrc and refreshing bash environment

```
# Place this to the .bashrc
export PATH="/home/swift/go/bin:$PATH"
```

Update the environment:
```
source ~/.bashrc
```