tasktray-proxy
===============

Simple Proxy Server on Tasktray for Windows and Linux (and Mac)

## Usage

### Step1 Create config file
`config.toml`

```toml
[[PROXY]]
description = "Your Proxy Name"
localHost   = "0.0.0.0"
localPort   = "8989"
authHost    = "some.company.auth.proxy"
authPort    = "8686"
user        = "[YOUR ID]"
password    = "[YOUR PASSWORD]"

[[PROXY]]
description = "Your Proxy Name 2"
localHost   = "0.0.0.0"
localPort   = "8989"

# ...
```

### Step2 Change Proxy Server
![Usage](./image/image.jpg)

## Build

### Windows(MinGW)
Build.command
```
go build -o ./tasktray-proxy.exe -ldflags -H=windowsgui main.go
```

### Linux
Require
```
sudo apt-get install libgtk-3-dev libappindicator3-dev
```
Build Command
```
go build -o ./tasktray-proxy.exe
```

### Mac
TODO
