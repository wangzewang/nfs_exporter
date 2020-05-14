package main

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/common/log"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type VolumeInfo struct {
	size     float64
	used     float64
	avail    float64
	capacity float64
}

type PathInfo struct {
	path    string
	used    float64
	pvName  string
	storage float64
}

type NfsPvInfo struct {
	server       string
	path         string
	capacity     string
	pvName       string
	pvcName      string
	pvcNamespace string
}

func execDfCommand(mountPath string) (*VolumeInfo, bool) {

	cmd := "df -k " + mountPath + " | grep dev | awk '{print $2\"-\"$3\"-\"$4\"-\"$5}'"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Errorf("df comand exec failed: %w", err.Error)
	}
	res := strings.Split(string(out), "-")
	size, err := strconv.ParseFloat(res[0], 64)
	if err != nil {
		log.Errorf("convert size error: %w", err.Error)
	}
	used, err := strconv.ParseFloat(res[1], 64)
	if err != nil {
		log.Errorf("convert used error: %w", err.Error)
	}
	avail, err := strconv.ParseFloat(res[2], 64)
	if err != nil {
		log.Errorf("convert avail error: %w", err.Error)
	}
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	capacity, err := strconv.ParseFloat(reg.ReplaceAllString(res[3], ""), 64)
	if err != nil {
		log.Errorf("convert capacity error: %w", err.Error)
	}

	return &VolumeInfo{
		size:     size * 1024,
		used:     used * 1024,
		avail:    avail * 1024,
		capacity: capacity,
	}, true
}

func execDuCommand(mountPath string) (*[]PathInfo, bool) {

	var pathsInfo []PathInfo
	cmd := "ls -l " + mountPath + "| awk '{print  \"" + mountPath + "/\"$9}' |  xargs -I {} du -shk \"{}\"| awk '{print $1\"@$@\"$2}'"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Errorf("du comand exec failed: %w", err)
	}

	res := strings.Split(string(out), "\n")
	log.Info(res)
	for _, v := range res {
		if v != "" {
			path := strings.Split(string(v), "@$@")[1]
			used, err := strconv.ParseFloat(strings.Split(v, "@$@")[0], 64)
			if err != nil {
				log.Errorf("convert avail error: %w", err.Error)
			}
			pathsInfo = append(pathsInfo,
				PathInfo{
					path: path,
					used: used * 1024,
				})
		}
	}

	return &pathsInfo, true
}

func getPvInfoFromCluster() map[string]map[string]NfsPvInfo {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	pvClient := clientset.CoreV1().PersistentVolumes()
	pvList, err := pvClient.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	pvInfo := make(map[string]map[string]NfsPvInfo)

	for _, pv := range pvList.Items {
		if pv.Status.Phase != v1.VolumeBound {
			continue
		}
		if pv.Spec.NFS != nil {
			pvInfo[pv.Spec.NFS.Server][pv.Spec.NFS.Server+pv.Spec.NFS.Path] = NfsPvInfo{
				server:       pv.Spec.NFS.Server,
				path:         pv.Spec.NFS.Path,
				capacity:     pv.Spec.Capacity.Storage().String(),
				pvName:       pv.Name,
				pvcName:      pv.Spec.ClaimRef.Name,
				pvcNamespace: pv.Spec.ClaimRef.Namespace,
			}
		}
	}

	return pvInfo
}
