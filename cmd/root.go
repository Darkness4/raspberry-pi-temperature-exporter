package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var (
	host    net.IP
	port    int
	sysfs   string
	rootCmd = &cobra.Command{
		Use:   "raspberry-pi-temperature-exporter",
		Short: "A temperature exporter for rasberry pi.",
		Run: func(cmd *cobra.Command, args []string) {
			go record()

			log.Printf("Serving at %s:%d\n", host, port)
			http.Handle("/metrics", promhttp.Handler())
			if err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil); err != nil {
				panic(err)
			}
		},
	}
	temperature = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "rpi",
		Subsystem: "thermal_zone0",
		Name:      "temp",
		Help:      "The temperature of the raspberry pi (°C)",
	})
)

func record() {
	for {
		err := func() error {
			b, err := ioutil.ReadFile(fmt.Sprintf("%s/class/thermal/thermal_zone0/temp", sysfs))
			if err != nil {
				return err
			}
			temp, err := strconv.ParseFloat(string(bytes.TrimSpace(b)), 64)
			if err != nil {
				return err
			}
			temp = temp / 1000.0
			temperature.Set(temp)
			return nil
		}()
		if err != nil {
			log.Printf("error: %+v\n", err)
		}
		time.Sleep(time.Second)
	}
}

func init() {
	rootCmd.PersistentFlags().IPVar(&host, "host", net.IPv4zero, "listening host")
	rootCmd.PersistentFlags().IntVar(&port, "port", 3000, "listening port")
	rootCmd.PersistentFlags().StringVar(&sysfs, "path.sysfs", "/sys", "/sys directory")
}

func Execute() error {
	return rootCmd.Execute()
}
