package main

import (
    "log"
    "os"
    "context"
    "github.com/things-go/go-socks5"
    "golang.org/x/net/proxy"
    "flag"
    "net"
    "net/http"
    "net/url"  // Add the url import
    "io"       // Add the io import
)

var target = flag.String("t", "1.1.1.1:1080", "proxy host")
var user = flag.String("u", "user", "username")
var password = flag.String("p", "password", "password")
var bind = flag.String("b", "127.0.0.1:1081", "bind address")
var proxyType = flag.String("type", "socks5", "proxy type (socks5 or http)")

func main() {
    flag.Parse()

    if *proxyType == "socks5" {
        startSocks5Proxy()
    } else if *proxyType == "http" {
        startHTTPProxy()
    } else {
        log.Fatalf("Unsupported proxy type: %s", *proxyType)
    }
}

func startSocks5Proxy() {
    dialSocksProxy, err := proxy.SOCKS5("tcp", *target, &proxy.Auth{User: *user, Password: *password}, proxy.Direct)
    if err != nil {
        log.Fatal(err)
    }

    ctxDialer := func(ctx context.Context, network, addr string) (net.Conn, error) {
        return dialSocksProxy.Dial(network, addr)
    }

    server := socks5.NewServer(
        socks5.WithLogger(socks5.NewLogger(log.New(os.Stdout, "socks5: ", log.LstdFlags))),
        socks5.WithDial(ctxDialer),
    )

    log.Printf("Starting SOCKS5 proxy %s://%s:%s@%s on %s://%s", *proxyType, *user, *password, *target,  *proxyType, *bind)
    if err := server.ListenAndServe("tcp", *bind); err != nil {
        log.Fatal(err)
    }
}

func startHTTPProxy() {
    // Create an HTTP transport for the proxy client (this can include the proxy auth)
    httpTransport := &http.Transport{
        Proxy: http.ProxyURL(&url.URL{
            Scheme: "http",
            Host:   *target,
            User:   url.UserPassword(*user, *password),
        }),
    }

    // Create a new HTTP proxy server handler
    httpProxyServer := &http.Server{
        Addr:    *bind,
        Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Forward the incoming HTTP request using the HTTP transport
            client := &http.Client{
                Transport: httpTransport,
            }

            // Copy the request's headers to the new request
            req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
            if err != nil {
                http.Error(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
                return
            }

            req.Header = r.Header.Clone()

            // Forward the request to the target proxy server
            resp, err := client.Do(req)
            if err != nil {
                http.Error(w, "Failed to forward request: "+err.Error(), http.StatusInternalServerError)
                return
            }
            defer resp.Body.Close()

            // Copy the response headers and body back to the original response writer
            for header, values := range resp.Header {
                for _, value := range values {
                    w.Header().Add(header, value)
                }
            }

            w.WriteHeader(resp.StatusCode)
            _, err = io.Copy(w, resp.Body)
            if err != nil {
                log.Printf("Failed to copy response body: %v", err)
            }
        }),
    }

    log.Printf("Starting HTTP proxy %s://%s:%s@%s on %s://%s", *proxyType, *user, *password, *target,  *proxyType, *bind)
    if err := httpProxyServer.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}
