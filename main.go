package main

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"./icon"
	"./proxyserver"
	"./proxyconfig"

	"github.com/getlantern/systray"
)

var (
	config *proxyconfig.Config
)

type ServerItem struct {
	server proxyserver.IProxy
	menu   *systray.MenuItem
}

func main() {
	config = proxyconfig.GetConfig("./config.toml")
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
		serverItem = append(serverItem, &ServerItem{p, m})
	}
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	serverItem = append(serverItem, &ServerItem{nil, mQuit})

	cases := make([]reflect.SelectCase, len(serverItem))
	for i, item := range serverItem {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(item.menu.ClickedCh)}
	}

	var proxy = serverItem[0].server
	proxy.Start()
	serverItem[0].menu.Check()
	changeTitle(0)

	go func() {
		for {
			chosen, _, _ := reflect.Select(cases)
			switch chosen {
			case len(cases) - 1:
				systray.Quit()
				os.Exit(0)
				return
			default:
				if !serverItem[chosen].menu.Checked() {
					changeTitle(chosen)
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

func changeTitle(index int) {
	str := "Tasktray Proxy(" + config.PROXY[index].LocalHost + ":" + config.PROXY[index].LocalPort + " =>" + config.PROXY[index].AuthHost + ":" + config.PROXY[index].AuthPort + ")"
	systray.SetTitle(str)
	systray.SetTooltip(str)
}

func uncheckWithout(serverItem []*ServerItem, index int) {
	for i := 0; i < len(serverItem); i++ {
		if i != index {
			serverItem[i].menu.Uncheck()
		}
	}
}
