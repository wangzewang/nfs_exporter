package main

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/common/log"
)

type VolumeInfo struct {
	size     float64
	used     float64
	avail    float64
	capacity float64
}

type PathInfo struct {
	path string
	used float64
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
	cmd := "ls -l " + mountPath + "| awk '{print  \"" + mountPath + "/\"$9}' |  xargs -I {} du -shk \"{}\"| awk '{print $1\"-\"$2}'"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Errorf("du comand exec failed: %w", err)
	}

	res := strings.Split(string(out), "\n")
	log.Info(res)
	for _, v := range res {
		if v != "" {
			path := strings.Split(string(v), "-")[1]
			used, err := strconv.ParseFloat(strings.Split(v, "-")[0], 64)
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
