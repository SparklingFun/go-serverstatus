package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	nativenet "net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

type StatusServer struct {
	Percent  StatusPercent
	Mem      MemInfo
	Swap     SwapInfo
	Load     *load.AvgStat
	Network  map[string]InterfaceInfo
	BootTime uint64
	Uptime   uint64
}
type BaseInfoServer struct {
	CPU    []CPUInfo
	System string
	IPAddr string
}
type StatusPercent struct {
	CPU  float64
	Disk float64
	Mem  float64
	Swap float64
}
type CPUInfo struct {
	ModelName string
	Cores     int32
}
type MemInfo struct {
	Total     uint64
	Used      uint64
	Available uint64
}
type SwapInfo struct {
	Total     uint64
	Used      uint64
	Available uint64
}
type InterfaceInfo struct {
	Addrs    []string
	ByteSent uint64
	ByteRecv uint64
}

// code below is modified from [Link](https://blog.zjyl1994.com/post/psutil/)
func main() {
	port := flag.String("port", ":19999", "HTTP listen port")
	flag.Parse()
	http.HandleFunc("/info", getBaseInfo)
	http.HandleFunc("/", getNowInfo)
	err := http.ListenAndServe(*port, nil)
	if err != nil {
		log.Fatalln("ListenAndServe: ", err)
	}
}
func getNowInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, infoMiniJSON())
}
func getBaseInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, baseMiniJson())
}

// Get preferred outbound ip of this machine
func GetOutboundIP() string {
	conn, err := nativenet.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*nativenet.UDPAddr)
	return localAddr.IP.String()
}

func baseMiniJson() string {
	ss := new(BaseInfoServer)
	c, _ := cpu.Info()
	n, _ := host.Info()
	ss.CPU = make([]CPUInfo, len(c))
	ss.System = strings.Join([]string{n.Platform, n.PlatformVersion, runtime.GOARCH, n.VirtualizationSystem}, "|")
	ss.IPAddr = GetOutboundIP()

	for i, ci := range c {
		ss.CPU[i].ModelName = ci.ModelName
		ss.CPU[i].Cores = ci.Cores
	}
	b, err := json.Marshal(ss)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}
func infoMiniJSON() string {
	v, _ := mem.VirtualMemory()
	s, _ := mem.SwapMemory()
	// c, _ := cpu.Info()
	cc, _ := cpu.Percent(time.Second, false)
	d, _ := disk.Usage("/")
	n, _ := host.Info()
	nv, _ := net.IOCounters(true)
	l, _ := load.Avg()
	// i, _ := net.Interfaces()
	ss := new(StatusServer)
	ss.Load = l
	ss.Uptime = n.Uptime
	ss.BootTime = n.BootTime
	ss.Percent.Mem = v.UsedPercent
	ss.Percent.CPU = cc[0]
	ss.Percent.Swap = s.UsedPercent
	ss.Percent.Disk = d.UsedPercent
	ss.Mem.Total = v.Total
	ss.Mem.Available = v.Available
	ss.Mem.Used = v.Used
	ss.Swap.Total = s.Total
	ss.Swap.Available = s.Free
	ss.Swap.Used = s.Used
	ss.Network = make(map[string]InterfaceInfo)
	// TODO: Network speed is currently 0 on WSL, but actually effect on Linux Servers
	for _, v := range nv {
		var ii InterfaceInfo
		ii.ByteSent = v.BytesSent
		ii.ByteRecv = v.BytesRecv
		ss.Network[v.Name] = ii
	}
	// for _, v := range i {
	// 	if ii, ok := ss.Network[v.Name]; ok {
	// 		ii.Addrs = make([]string, len(v.Addrs))
	// 		for i, vv := range v.Addrs {
	// 			ii.Addrs[i] = vv.Addr
	// 		}
	// 		ss.Network[v.Name] = ii
	// 	}
	// }
	b, err := json.Marshal(ss)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}
