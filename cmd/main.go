package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/didikz/gosysmon/internal/server"
	"github.com/didikz/gosysmon/pkg/util"
	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

func main() {
	fmt.Println("Starting system monitor")
	s := server.NewHttpServer()

	go func(s *server.HttpServer) {
		for {
			hostStat, _ := host.Info()
			vmStat, _ := mem.VirtualMemory()

			cpus, _ := cpu.Info()
			var cpuInfo string
			for _, cpu := range cpus {
				cpuInfo += `
				<li class="flex justify-between gap-x-4 py-1 bg-gray-500 rounded-sm">
					<span class="mx-2 p-1">Manufacturer</span>
					<span class="mx-2 p-1" id="cpu-manufacturer">` + cpu.VendorID + `</span>
				</li>
				<li class="flex justify-between gap-x-4 py-1 rounded-sm">
					<span class="mx-2 p-1">Model</span>
					<span class="mx-2 p-1" id="cpu-model">` + cpu.ModelName + `</span>
				</li>
				<li class="flex justify-between gap-x-4 py-1 bg-gray-500 rounded-sm">
					<span class="mx-2 p-1">Family</span>
					<span class="mx-2 p-1" id="cpu-family">` + cpu.Family + `</span>
				</li>
				<li class="flex justify-between gap-x-4 py-1 rounded-sm">
					<span class="mx-2 p-1">Speed</span>
					<span class="mx-2 p-1" id="cpu-speed">` + fmt.Sprintf("%.2f MHz", cpu.Mhz) + `</span>
				</li>
				<li class="flex justify-between gap-x-4 py-1 bg-gray-500 rounded-sm">
					<span class="mx-2 p-1">Cores</span>
					<span class="mx-2 p-1" id="cpu-cores">` + fmt.Sprintf("%d cores", cpu.Cores) + `</span>
				</li>
				`
			}

			partitionStats, _ := disk.Partitions(true)
			partitions, totalStorage, usedStorage, freeStorage := "", "", "", ""

			for _, partition := range partitionStats {
				diskUsage, _ := disk.Usage(partition.Mountpoint)
				if partitions == "" {
					partitions = fmt.Sprintf("%s (%s)", partition.Mountpoint, partition.Fstype)
				} else {
					partitions += fmt.Sprintf(", %s (%s)", partition.Mountpoint, partition.Fstype)
				}

				if totalStorage == "" {
					totalStorage = fmt.Sprintf("%s %dGB", partition.Mountpoint, util.BytesToGigabyte(diskUsage.Total))
				} else {
					totalStorage += fmt.Sprintf(", %s %dGB", partition.Mountpoint, util.BytesToGigabyte(diskUsage.Total))
				}

				if usedStorage == "" {
					usedStorage = fmt.Sprintf("%s %dGB (%.2f%%)", partition.Mountpoint, util.BytesToGigabyte(diskUsage.Used), diskUsage.UsedPercent)
				} else {
					usedStorage += fmt.Sprintf(", %s %dGB (%.2f%%)", partition.Mountpoint, util.BytesToGigabyte(diskUsage.Used), diskUsage.UsedPercent)
				}

				if freeStorage == "" {
					freeStorage = fmt.Sprintf("%s %dGB", partition.Mountpoint, util.BytesToGigabyte(diskUsage.Free))
				} else {
					freeStorage += fmt.Sprintf(", %s %dGB", partition.Mountpoint, util.BytesToGigabyte(diskUsage.Free))
				}
			}

			timestamp := time.Now().Format("2006-01-02 15:04:05")
			html := `
			<span hx-swap-oob="innerHTML:#data-timestamp">` + timestamp + `</span>
			<span hx-swap-oob="innerHTML:#system-hostname">` + hostStat.Hostname + `</span>
			<span hx-swap-oob="innerHTML:#system-os">` + hostStat.OS + `</span>
			<span hx-swap-oob="innerHTML:#system-platform">` + fmt.Sprintf("%s (%s)", hostStat.Platform, hostStat.PlatformFamily) + `</span>
			<span hx-swap-oob="innerHTML:#system-version">` + hostStat.PlatformVersion + `</span>
			<span hx-swap-oob="innerHTML:#system-arch">` + fmt.Sprintf("%s (%s)", hostStat.KernelArch, hostStat.KernelVersion) + `</span>
			<span hx-swap-oob="innerHTML:#system-running-processess">` + strconv.Itoa(int(hostStat.Procs)) + `</span>
			<span hx-swap-oob="innerHTML:#system-total-memory">` + strconv.Itoa(int(util.BytesToGigabyte(vmStat.Total))) + `GB </span>
			<span hx-swap-oob="innerHTML:#system-used-memory">` + strconv.Itoa(int(util.BytesToGigabyte(vmStat.Used))) + `GB (` + fmt.Sprintf("%.2f%%", vmStat.UsedPercent) + `)</span>
			<span hx-swap-oob="innerHTML:#system-free-memory">` + strconv.Itoa(int(util.BytesToGigabyte(vmStat.Free))) + `GB </span>
			<div hx-swap-oob="innerHTML:#cpu-data">` + cpuInfo + `</div>
			<span hx-swap-oob="innerHTML:#disk-partitions">` + partitions + `</span>
			<span hx-swap-oob="innerHTML:#disk-total-storage">` + totalStorage + `</span>
			<span hx-swap-oob="innerHTML:#disk-usage">` + usedStorage + `</span>
			<span hx-swap-oob="innerHTML:#disk-free">` + freeStorage + `</span>
			`
			s.Broadcast([]byte(html))
			time.Sleep(time.Second * 3)
		}
	}(s)

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	err = http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("APP_PORT")), &s.Mux)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
