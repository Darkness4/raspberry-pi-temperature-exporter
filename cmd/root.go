package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
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
	gauges  map[string]prometheus.Gauge
	rootCmd = &cobra.Command{
		Use:   "raspberry-pi-temperature-exporter",
		Short: "A temperature exporter for rasberry pi.",
		Run: func(cmd *cobra.Command, args []string) {
			gauges = make(map[string]prometheus.Gauge)
			dir := fmt.Sprintf("%s/class/thermal", sysfs)

			// Find thermal sensors
			infos, err := ioutil.ReadDir(dir)
			if err != nil {
				panic(err)
			}

			// Filter by "thermal_zone*"
			for _, info := range infos {
				if strings.HasPrefix(info.Name(), "thermal_zone") {
					log.Printf("found %s\n", info.Name())
					b, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/type", dir, info.Name()))
					if err != nil {
						panic(err)
					}

					gauges[info.Name()] = promauto.NewGauge(prometheus.GaugeOpts{
						Namespace: "linux",
						Subsystem: "thermal_zone",
						Name:      "temp",
						Help:      "Temperature indicated by a linux device (°C)",
						ConstLabels: prometheus.Labels{
							"sensor": info.Name(),
							"type":   string(bytes.TrimSpace(b)),
						},
					})
				}
			}

			go record()

			log.Printf("Serving at %s:%d\n", host, port)
			http.Handle("/metrics", promhttp.Handler())
			if err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil); err != nil {
				panic(err)
			}
		},
	}
)

func record() {
	thermaldir := fmt.Sprintf("%s/class/thermal", sysfs)
	for {
		// Parallel calls to the sensors
		var wait sync.WaitGroup
		for name, gauge := range gauges {
			wait.Add(1)
			go func(name string, gauge prometheus.Gauge) {
				defer wait.Done()
				err := func() error {
					b, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/temp", thermaldir, name))
					if err != nil {
						return err
					}
					temp, err := strconv.ParseFloat(string(bytes.TrimSpace(b)), 64)
					if err != nil {
						return err
					}
					temp = temp / 1000.0
					gauge.Set(temp)
					return nil
				}()
				if err != nil {
					log.Printf("error: %+v\n", err)
				}
			}(name, gauge)
		}
		time.Sleep(time.Second)
		wait.Wait()
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
