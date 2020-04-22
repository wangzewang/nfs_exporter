package main

import (
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	volumeSize = prometheus.NewDesc(prometheus.BuildFQName("", "", "nfs_total_size"),
		"NFS total size", []string{"server", "mount_path"}, nil)
	volumeUsed = prometheus.NewDesc(prometheus.BuildFQName("", "", "nfs_used_size"),
		"NFS total used size", []string{"server", "mount_path"}, nil)
	volumeAvail = prometheus.NewDesc(prometheus.BuildFQName("", "", "nfs_free_size"),
		"NFS free size", []string{"server", "mount_path"}, nil)
	volumeCapacity = prometheus.NewDesc(prometheus.BuildFQName("", "", "nfs_capacity"),
		"NFS capacity", []string{"server", "mount_path"}, nil)
	volumeDataUsed = prometheus.NewDesc(prometheus.BuildFQName("", "", "nfs_volume_used_size"),
		"NFS volume used size", []string{"server", "mount_path", "path"}, nil)
)

// Exporter holds name, path and volumes to be monitored
type Exporter struct {
	nfsServer string
	nfsPath   string
}

// Describe all the metrics exported by NFS exporter. It implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- volumeSize
	ch <- volumeUsed
	ch <- volumeAvail
	ch <- volumeCapacity
}

// Collect collects all the metrics
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	volumeInfo, _ := execDfCommand(e.nfsPath)
	volumeDataInfo, _ := execDuCommand(e.nfsPath)
	ch <- prometheus.MustNewConstMetric(volumeSize, prometheus.GaugeValue, volumeInfo.size, e.nfsServer, e.nfsPath)
	ch <- prometheus.MustNewConstMetric(volumeUsed, prometheus.GaugeValue, volumeInfo.used, e.nfsServer, e.nfsPath)
	ch <- prometheus.MustNewConstMetric(volumeAvail, prometheus.GaugeValue, volumeInfo.avail, e.nfsServer, e.nfsPath)
	ch <- prometheus.MustNewConstMetric(volumeCapacity, prometheus.GaugeValue, volumeInfo.capacity, e.nfsServer, e.nfsPath)
	for _, v := range *volumeDataInfo {
		ch <- prometheus.MustNewConstMetric(volumeDataUsed, prometheus.GaugeValue, v.used, e.nfsServer, e.nfsPath, v.path)

	}
}

// NewExporter initialises exporter
func NewExporter(nfsPath, nfsServer string) (*Exporter, error) {
	return &Exporter{
		nfsServer: nfsServer,
		nfsPath:   nfsPath,
	}, nil
}

func init() {
	prometheus.MustRegister(version.NewCollector("nfs_exporter"))
}

func main() {
	var (
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		listenAddress = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9102").String()
		nfsPath       = kingpin.Flag("nfs.storage-path", "Path to nfs storage volume.").Default("/tmp/nfs").String()
		nfsServer     = kingpin.Flag("nfs.server", "IP address to nfs storage cluster.").Default("127.0.0.1").String()
	)
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("nfs_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting nfs_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	exporter, err := NewExporter(*nfsPath, *nfsServer)
	if err != nil {
		log.Fatalf("Creating new Exporter went wrong, ... \n%v", err)
	}
	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>NFS Exporter</title></head>
             <body>
			<h1>NFS Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		if err != nil {
			log.Fatal("Error starting HTTP server", err)
			os.Exit(1)
		}
	}

}
