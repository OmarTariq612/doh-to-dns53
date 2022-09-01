package main

import (
	"encoding/base64"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

const (
	timeoutDuration     = 5 * time.Second
	maxDNSMessageLength = 512
	dnsServerAddr       = "8.8.8.8:53"
	certificate         = "site.crt"
	key                 = "site.key"
)

func main() {
	srv := &http.Server{
		Addr:              ":443",
		Handler:           http.TimeoutHandler(http.DefaultServeMux, 10*time.Second, ""),
		IdleTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	http.HandleFunc("/dns-query", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			_, _ = io.Copy(ioutil.Discard, r.Body)
			_ = r.Body.Close()
		}()

		conn, err := net.Dial("udp", dnsServerAddr)
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		var buf [maxDNSMessageLength]byte

		switch r.Method {
		case http.MethodGet:
			log.Println("GET")
			requestBase64UrlEncoded := r.URL.Query()["dns"]
			n, _ := base64.RawURLEncoding.Decode(buf[:], []byte(requestBase64UrlEncoded[0]))
			_, _ = conn.Write(buf[:n])
			conn.SetReadDeadline(time.Now().Add(timeoutDuration))
			n, _ = conn.Read(buf[:])
			w.Header().Add("Content-Type", "application/dns-message")
			w.Write(buf[:n])
		case http.MethodPost:
			log.Println("POST")
			n, _ := r.Body.Read(buf[:])
			_, _ = conn.Write(buf[:n])
			conn.SetReadDeadline(time.Now().Add(timeoutDuration))
			n, _ = conn.Read(buf[:])
			w.Header().Add("Content-Type", "application/dns-message")
			_, _ = w.Write(buf[:n])
		}
	})

	log.Println(srv.ListenAndServeTLS("site.crt", "site.key"))
}
