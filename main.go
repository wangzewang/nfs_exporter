package main

import (
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
	"strings"
	"time"
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
	volumeDataUsedWithPv = prometheus.NewDesc(prometheus.BuildFQName("", "", "nfs_volume_used_size"),
		"NFS volume used size", []string{"server", "mount_path", "path", "pv_name", "pv_total_size", "pvc_name", "pvc_namespace"}, nil)
)

// Exporter holds name, path and volumes to be monitored
type Exporter struct {
	nfsServerPaths string
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
	var pvInfo map[string]map[string]NfsPvInfo
	if x, found := dataCache.Get("pvInfo"); found {
		pvInfo = x.(map[string]map[string]NfsPvInfo)
	}

	for _, serverPath := range strings.Split(e.nfsServerPaths, ",") {
		var volumeInfo VolumeInfo
		var volumeDataInfo []PathInfo

		ip := strings.Split(serverPath, ":")[0]
		path := strings.Split(serverPath, ":")[1]
		if x, found := dataCache.Get(ip + "volumeInfo"); found {
			volumeInfo = x.(VolumeInfo)
		}
		if x, found := dataCache.Get(ip + "volumeDataInfo"); found {
			volumeDataInfo = x.([]PathInfo)
		}
		ch <- prometheus.MustNewConstMetric(volumeSize, prometheus.GaugeValue, volumeInfo.size, ip, path)
		ch <- prometheus.MustNewConstMetric(volumeUsed, prometheus.GaugeValue, volumeInfo.used, ip, path)
		ch <- prometheus.MustNewConstMetric(volumeAvail, prometheus.GaugeValue, volumeInfo.avail, ip, path)
		ch <- prometheus.MustNewConstMetric(volumeCapacity, prometheus.GaugeValue, volumeInfo.capacity, ip, path)
		for _, val := range volumeDataInfo {
			if v, exist := pvInfo[ip][ip+val.path]; exist {
				ch <- prometheus.MustNewConstMetric(volumeDataUsedWithPv, prometheus.GaugeValue, val.used, ip, path, val.path, v.pvName, v.capacity, v.pvcName, v.pvcNamespace)
			} else {
				ch <- prometheus.MustNewConstMetric(volumeDataUsed, prometheus.GaugeValue, val.used, ip, path, val.path)
			}
		}

	}
}

// NewExporter initialises exporter
func NewExporter(nfsServerPaths string) (*Exporter, error) {
	return &Exporter{
		nfsServerPaths: nfsServerPaths,
	}, nil
}

func collectData(nfsServerPaths string) {
	for {
		dataCache.Set("pvInfo", getPvInfoFromCluster(), cache.NoExpiration)
		for _, serverPath := range strings.Split(nfsServerPaths, ",") {
			ip := strings.Split(serverPath, ":")[0]
			path := strings.Split(serverPath, ":")[1]
			dataCache.Set(ip+"volumeInfo", execDfCommand(path), cache.NoExpiration)
			dataCache.Set(ip+"volumeDataInfo", execDuCommand(path), cache.NoExpiration)
		}
		time.Sleep(5 * time.Minute)
	}
}

var dataCache *cache.Cache

func main() {
	var (
		metricsPath    = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		listenAddress  = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9102").String()
		nfsServerPaths = kingpin.Flag("nfs.address", "Nfs servers ip with patha, split with comma. Example: 192.168.1.1:/data, 192.168.1.2:/data2").Default(os.Getenv("NFS_SERVER_PATHS")).String()
	)
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("nfs_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	dataCache = cache.New(cache.NoExpiration, cache.NoExpiration)

	go collectData(*nfsServerPaths)

	log.Infoln("Starting nfs_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	exporter, err := NewExporter(*nfsServerPaths)
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
