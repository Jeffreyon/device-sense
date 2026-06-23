//go:build !windows

package main

import (
	"os/exec"
	"runtime"
	"strings"
)

func enableAnsiColors() {} // ANSI works natively on Linux/macOS

func run(name string, args ...string) string {
	out, _ := exec.Command(name, args...).Output()
	return strings.TrimSpace(string(out))
}

func scanCPU() CPUResult {
	var name, cores, threads, speed string

	switch runtime.GOOS {
	case "linux":
		raw := run("cat", "/proc/cpuinfo")
		for _, line := range strings.Split(raw, "\n") {
			if strings.HasPrefix(line, "model name") && name == "" {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					name = strings.TrimSpace(parts[1])
				}
			}
			if strings.HasPrefix(line, "cpu cores") && cores == "" {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					cores = strings.TrimSpace(parts[1])
				}
			}
			if strings.HasPrefix(line, "siblings") && threads == "" {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					threads = strings.TrimSpace(parts[1])
				}
			}
			if strings.HasPrefix(line, "cpu MHz") && speed == "" {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					speed = strings.TrimSpace(parts[1])
				}
			}
		}
	case "darwin":
		name = run("sysctl", "-n", "machdep.cpu.brand_string")
		cores = run("sysctl", "-n", "hw.physicalcpu")
		threads = run("sysctl", "-n", "hw.logicalcpu")
		speed = run("sysctl", "-n", "hw.cpufrequency_max")
	}

	return CPUResult{
		RegistryName: "N/A (registry check is Windows-only)",
		WMIName:      name,
		Cores:        cores,
		Threads:      threads,
		SpeedMHz:     speed,
		Arch:         runtime.GOARCH,
	}
}

func scanRAM() ([]RAMSlot, string) {
	var total string
	switch runtime.GOOS {
	case "linux":
		raw := run("grep", "MemTotal", "/proc/meminfo")
		total = raw
	case "darwin":
		total = run("sysctl", "-n", "hw.memsize")
	}
	return []RAMSlot{{Slot: "1", CapacityGB: total, Type: "See dmidecode for details"}}, total
}

func scanStorage() []Disk {
	var out string
	switch runtime.GOOS {
	case "linux":
		out = run("lsblk", "-d", "-o", "NAME,SIZE,TYPE,MODEL")
	case "darwin":
		out = run("diskutil", "list")
	}
	return []Disk{{Caption: out}}
}

func scanGPU() []GPU {
	var name string
	switch runtime.GOOS {
	case "linux":
		name = run("lspci", "-v")
	case "darwin":
		name = run("system_profiler", "SPDisplaysDataType")
	}
	return []GPU{{Name: name}}
}

func scanSystem() SysInfo {
	switch runtime.GOOS {
	case "linux":
		mfr := run("cat", "/sys/class/dmi/id/sys_vendor")
		model := run("cat", "/sys/class/dmi/id/product_name")
		bios := run("cat", "/sys/class/dmi/id/bios_version")
		serial := run("cat", "/sys/class/dmi/id/product_serial")
		return SysInfo{Manufacturer: mfr, Model: model, BIOSVersion: bios, Serial: serial}
	case "darwin":
		info := run("system_profiler", "SPHardwareDataType")
		return SysInfo{Model: info}
	}
	return SysInfo{}
}
