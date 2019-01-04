package main

import (
	"flag"
	"fmt"
	"github.com/gashirar/tasktray-proxy/icon"
	"github.com/gashirar/tasktray-proxy/proxyconfig"
	"github.com/gashirar/tasktray-proxy/proxyserver"
	"github.com/gashirar/tasktray-proxy/wifiname"
	"github.com/getlantern/systray"
	"log"
	"net"
	"os"
	"reflect"
	"time"
)

var (
	config *proxyconfig.Config
)

type ServerItem struct {
	config proxyconfig.ProxyConfig
	server proxyserver.IProxy
	menu   *systray.MenuItem
}

func main() {
	strOpt := flag.String("c", "./config.toml", "help message for s option")
	flag.Parse()
	fmt.Println("config file : ", *strOpt)
	var err error
	config, err = proxyconfig.GetConfig(*strOpt)
	if err != nil {
		log.Fatal(err)
	}
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(icon.Data)
	serverItem := []*ServerItem{}

	for _, cnf := range config.PROXY {
		m := systray.AddMenuItem(cnf.Description, cnf.Description)
		var p proxyserver.IProxy
		if cnf.AuthPort != "" && cnf.AuthHost != "" && cnf.User != "" && cnf.Password != "" {
			p = proxyserver.NewAuthProxy(cnf.LocalHost, cnf.LocalPort, cnf.AuthHost, cnf.AuthPort, cnf.User, cnf.Password)
		} else {
			p = proxyserver.NewProxy(cnf.LocalHost, cnf.LocalPort)
		}
		serverItem = append(serverItem, &ServerItem{cnf, p, m})
	}
	systray.AddSeparator()
	autoSwitch := systray.AddMenuItem("Auto Switch", "Auto Switch Proxy")
	serverItem = append(serverItem, &ServerItem{proxyconfig.ProxyConfig{Description: "Auto Change"}, nil, autoSwitch})
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	serverItem = append(serverItem, &ServerItem{proxyconfig.ProxyConfig{}, nil, mQuit})

	cases := make([]reflect.SelectCase, len(serverItem))
	for i, item := range serverItem {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(item.menu.ClickedCh)}
	}

	var proxy = serverItem[0].server
	proxy.Start()
	serverItem[0].menu.Check()
	changeTitle(serverItem[0].config.Description)

	if config.AutoSwitch {
		autoSwitch.Check()
		autoSwitch.SetTitle("✓Auto Switch")
	}

	go func() {
		t := time.NewTicker(5 * time.Second) // 3秒おきに通知
		for {
			select {
			case <-t.C:
				if !autoSwitch.Checked() {
					continue
				}
				name := wifiname.WifiName()
				ipAddrs := getIpAddresses()
				for _, item := range serverItem {
					if name != "" && name == item.config.Wifi {
						item.menu.ClickedCh <- struct{}{}
					} else if item.config.Network != "" {
						for _, ip := range ipAddrs {
							if isIncluededInCIDR(item.config.Network, ip) {
								item.menu.ClickedCh <- struct{}{}
							}
						}
					}
				}
			}
		}
		t.Stop()
	}()

	go func() {
		for {
			chosen, _, _ := reflect.Select(cases)
			switch chosen {
			case len(cases) - 1:
				systray.Quit()
				os.Exit(0)
				return
			case len(cases) - 2:
				if !autoSwitch.Checked() {
					autoSwitch.Check()
					autoSwitch.SetTitle("✓" + serverItem[chosen].config.Description)
				} else {
					autoSwitch.Uncheck()
					autoSwitch.SetTitle(serverItem[chosen].config.Description)
				}
			default:
				if !serverItem[chosen].menu.Checked() {
					changeTitle(serverItem[chosen].config.Description)
					proxy = changeProxyServer(proxy, serverItem, chosen)
				}
			}
		}
	}()
}

func onExit() {
	fmt.Println("Finished onExit")
}

func changeProxyServer(currentProxy proxyserver.IProxy, serverItem []*ServerItem, index int) proxyserver.IProxy {
	currentProxy.Shutdown()
	time.Sleep(3 * time.Second)
	serverItem[index].server.Start()
	uncheckWithout(serverItem, index)
	serverItem[index].menu.Check()
	return serverItem[index].server
}

func changeTitle(title string) {
	systray.SetTitle(title)
	systray.SetTooltip(title)
}

func uncheckWithout(serverItem []*ServerItem, index int) {
	for i := 0; i < len(serverItem)-2; i++ {
		if i != index {
			serverItem[i].menu.Uncheck()
		}
	}
}

func getIpAddresses() []string {
	addrs, err := net.InterfaceAddrs()
	var ips []string
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		return ips
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	return ips
}

func isIncluededInCIDR(cidr string, ip string) bool {
	_, cidrNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	targetIP := net.ParseIP(ip)
	return cidrNet.Contains(targetIP)
}
