package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	replyString            = strings.Repeat("HelloWorld", 25) + "\r\n"
	sshListenerFlag        = flag.String("ssh", ":22", "listen address for ssh connections")
	prometheusListenerFlag = flag.String("prom", "127.0.0.1:5000", "listen address for prometheus http")
)

func main() {
	flag.Parse()

	log.Printf("listening for SSH connections on %q...", *sshListenerFlag)
	lis, err := net.Listen("tcp4", *sshListenerFlag)
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
			log.Printf("accept error: %v", err)
			continue
		}

		go func() {
			started := time.Now()

			log.Printf("handling %v...", conn.RemoteAddr().String())
			activeConnectionsCount.Inc()
			defer func() {
				activeConnectionsCount.Dec()
				totalConnectionsCount.Inc()
				connectionLifetime.Set(time.Since(started).Seconds())
			}()

			for {
				time.Sleep(1 * time.Second)
				sent, err := conn.Write([]byte(replyString))
				if err != nil {
					log.Printf("client %s got error %v, stop writing...", conn.RemoteAddr().String(), err)
					return
				}
				connectionBytesSent.Add(float64(sent))
			}
		}()
	}
}

const (
	promNamespace = "tarpit"
	promSubsystem = "ssh"
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
