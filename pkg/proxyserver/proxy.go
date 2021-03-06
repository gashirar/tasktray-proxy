package proxyserver

import (
	"net/http"
	"fmt"
	"log"
	"time"
	"context"
	"net/url"
	"net"
	"io"
	"encoding/base64"
)

type IProxy interface {
	Start() error
	Shutdown() error
}

type Proxy struct {
	server    *http.Server
	handler   http.Handler
	transport http.RoundTripper
	host      string
	port      string
}

func (p *Proxy) Start() (err error) {
	p.server = &http.Server{
		Addr:    p.host + ":" + p.port,
		Handler: p.handler,
	}

	go func() {
		http.DefaultTransport = p.transport
		listen := fmt.Sprintf("%s:%s", p.host, p.port)
		log.Printf("Start Proxy serving on %s", listen)
		if err = p.server.ListenAndServe(); err != nil {
			log.Printf("Start Proxy Server: %s", err)
		}
	}()
	return
}

func (p *Proxy) Shutdown() (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err = p.server.Shutdown(ctx)
	return
}

func (p *Proxy) handleHttps(w http.ResponseWriter, r *http.Request) {

	destConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		log.Print(err)
	}
	w.WriteHeader(http.StatusOK)
	hj, _ := w.(http.Hijacker)
	if clientConn, _, err := hj.Hijack(); err != nil {
		destConn.Close()
		log.Print(err)
	} else {
		go transfer(destConn, clientConn)
		go transfer(clientConn, destConn)
	}
}

func (p *Proxy) handleHttp(w http.ResponseWriter, r *http.Request) {
	hj, _ := w.(http.Hijacker)
	client := &http.Client{}
	r.RequestURI = ""
	if resp, err := client.Do(r); err != nil {
		log.Print(err)
	} else if conn, _, err := hj.Hijack(); err != nil {
		log.Print(err)
	} else {
		defer conn.Close()
		defer resp.Body.Close()
		resp.Write(conn)
	}
}

func (p *Proxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] %s", r.Method, r.URL)
	if r.Method == "CONNECT" {
		p.handleHttps(w, r)
	} else {
		p.handleHttp(w, r)
	}
}

type AuthProxy struct {
	*Proxy
	authHost           string
	authPort           string
	username           string
	password           string
	proxyAuthorization string
}

func (p *AuthProxy) proxyTransport() http.RoundTripper {
	proxyUrlString := fmt.Sprintf("http://%s:%s@%s:%s", url.QueryEscape(p.username), url.QueryEscape(p.password), p.authHost, p.authPort)
	proxyUrl, err := url.Parse(proxyUrlString)
	if err != nil {
		log.Fatal(err)
	}
	return &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
}

func (p *AuthProxy) handleHttps(w http.ResponseWriter, r *http.Request) {
	hj, _ := w.(http.Hijacker)

	if proxyConn, err := net.Dial("tcp", p.authHost+":"+p.authPort); err != nil {
		log.Print(err)
	} else if clientConn, _, err := hj.Hijack(); err != nil {
		proxyConn.Close()
		log.Print(err)
	} else {
		r.Header.Set("Proxy-Authorization", p.proxyAuthorization)
		r.Write(proxyConn)
		go transfer(proxyConn, clientConn)
		go transfer(clientConn, proxyConn)
	}
}

func (p *AuthProxy) handleHttp(w http.ResponseWriter, r *http.Request) {
	hj, _ := w.(http.Hijacker)
	client := &http.Client{}
	r.RequestURI = ""
	if resp, err := client.Do(r); err != nil {
		log.Print(err)
	} else if conn, _, err := hj.Hijack(); err != nil {
		log.Print(err)
	} else {
		defer conn.Close()
		defer resp.Body.Close()
		resp.Write(conn)
	}
}

func (p *AuthProxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] %s", r.Method, r.URL)
	if r.Method == "CONNECT" {
		p.handleHttps(w, r)
	} else {
		p.handleHttp(w, r)
	}
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer func() {
		if destination != nil {
			destination.Close()
		}
	}()
	defer func() {
		if source != nil {
			source.Close()
		}
	}()
	if destination != nil && source != nil {
		io.Copy(destination, source)
	}
}

func NewProxy(host string, port string) *Proxy {
	p := &Proxy{host: host, port: port}

	p.transport = &http.Transport{
		Proxy: nil,
	}

	p.handler = http.HandlerFunc(p.handleRequest)

	return p
}

func NewAuthProxy(host string, port string, authHost string, authPort string, username string, password string) *AuthProxy {
	p := &AuthProxy{Proxy: &Proxy{host: host, port: port}, authHost: authHost, authPort: authPort, username: username, password: password}

	p.transport = p.proxyTransport()
	p.proxyAuthorization = "Basic " + base64.StdEncoding.EncodeToString([]byte(p.username+":"+p.password))

	p.handler = http.HandlerFunc(p.handleRequest)

	return p
}
