package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpListenerFlag       = flag.String("http", ":80", "listen address for HTTP connections")
	prometheusListenerFlag = flag.String("prom", "127.0.0.1:5000", "listen address for prometheus http")
)

func main() {
	flag.Parse()

	log.Printf("listening for HTTP connections on %q...", *httpListenerFlag)
	lis, err := net.Listen("tcp4", *httpListenerFlag)
	if err != nil {
		panic(err)
	}

	go func() {
		log.Printf("starting metrics server on %q...", *prometheusListenerFlag)
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(*prometheusListenerFlag, nil); err != nil {
			panic(err)
		}
	}()

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Printf("accept failed: %v", err)
			continue
		}

		go func() {
			started := time.Now()
			activeConnectionsCount.Inc()

			log.Printf("handling %q...", conn.RemoteAddr().String())

			defer func() {
				activeConnectionsCount.Dec()
				totalConnectionsCount.Inc()
				connectionLifetime.Set(time.Since(started).Seconds())
			}()

			n, err := conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
			if err != nil {
				log.Printf("failed to write to %q: %v", conn.RemoteAddr().String(), err)
				return
			}
			connectionBytesSent.Add(float64(n))

			for {
				time.Sleep(1 * time.Second)
				header, value := rand.Int63(), rand.Int63()
				n, err = conn.Write([]byte(fmt.Sprintf("X-%x: %x\r\n", header, value)))
				if err != nil {
					log.Printf("failed to write to %q: %v", conn.RemoteAddr().String(), err)
					return
				}
				connectionBytesSent.Add(float64(n))
			}
		}()
	}
}

const (
	promNamespace = "tarpit"
	promSubsystem = "http"
)

var (
	totalConnectionsCount = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "conn_total_count",
	})
	activeConnectionsCount = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "conn_active_count",
	})
	connectionLifetime = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "conn_lifetime",
	})
	connectionBytesSent = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "conn_bytes_sent",
	})
)
