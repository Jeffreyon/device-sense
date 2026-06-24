//go:build !windows

package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func enableAnsiColors() {}

func run(name string, args ...string) string {
	out, _ := exec.Command(name, args...).Output()
	return strings.TrimSpace(string(out))
}

func field(raw, key string) string {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), strings.ToLower(key)) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

func scanCPU() CPUResult {
	var name, cores, threads, speedMHz, arch string

	switch runtime.GOOS {
	case "linux":
		raw := run("cat", "/proc/cpuinfo")
		name = field(raw, "model name")
		cores = field(raw, "cpu cores")
		threads = field(raw, "siblings")
		speedMHz = field(raw, "cpu mhz")
		arch = run("uname", "-m")

	case "darwin":
		name = run("sysctl", "-n", "machdep.cpu.brand_string")
		cores = run("sysctl", "-n", "hw.physicalcpu")
		threads = run("sysctl", "-n", "hw.logicalcpu")
		hz := run("sysctl", "-n", "hw.cpufrequency_max")
		if hz != "" {
			// convert Hz to MHz
			var hzInt int64
			fmt.Sscanf(hz, "%d", &hzInt)
			speedMHz = fmt.Sprintf("%.0f", float64(hzInt)/1_000_000)
		}
		arch = run("uname", "-m")
	}

	speedGHz := ""
	if speedMHz != "" {
		var mhz float64
		fmt.Sscanf(speedMHz, "%f", &mhz)
		speedGHz = fmt.Sprintf("%.2f", mhz/1000)
	}

	return CPUResult{
		RegistryName: "N/A — registry check is Windows-only",
		WMIName:      name,
		Cores:        cores,
		Threads:      threads,
		SpeedMHz:     speedMHz,
		SpeedGHz:     speedGHz,
		Arch:         arch,
	}
}

func scanRAM() ([]RAMSlot, string) {
	switch runtime.GOOS {
	case "linux":
		raw := run("cat", "/proc/meminfo")
		totalKB := field(raw, "MemTotal")
		totalGB := ""
		if totalKB != "" {
			var kb int64
			fmt.Sscanf(totalKB, "%d", &kb)
			totalGB = fmt.Sprintf("%.2f", float64(kb)/1_048_576)
		}
		return []RAMSlot{}, totalGB

	case "darwin":
		bytes := run("sysctl", "-n", "hw.memsize")
		totalGB := ""
		if bytes != "" {
			var b int64
			fmt.Sscanf(bytes, "%d", &b)
			totalGB = fmt.Sprintf("%.2f", float64(b)/1_073_741_824)
		}
		return []RAMSlot{}, totalGB
	}
	return []RAMSlot{}, "N/A"
}

func scanStorage() []Disk {
	switch runtime.GOOS {
	case "linux":
		raw := run("lsblk", "-d", "-o", "NAME,SIZE,TYPE,MODEL", "--noheadings")
		var disks []Disk
		for _, line := range strings.Split(raw, "\n") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			name := fields[0]
			size := fields[1]
			model := ""
			if len(fields) >= 4 {
				model = strings.Join(fields[3:], " ")
			}
			disks = append(disks, Disk{
				Caption:   fmt.Sprintf("%s (%s)", name, model),
				SizeGB:    size,
				Interface: "—",
				Serial:    "run: udevadm info /dev/" + name,
			})
		}
		return disks

	case "darwin":
		raw := run("diskutil", "list", "-plist")
		if raw == "" {
			raw = run("diskutil", "list")
		}
		return []Disk{{Caption: "Run  diskutil list  for full details", SizeGB: "—", Interface: "—", Serial: "—"}}
	}
	return nil
}

func scanGPU() []GPU {
	switch runtime.GOOS {
	case "linux":
		raw := run("lspci")
		var gpus []GPU
		for _, line := range strings.Split(raw, "\n") {
			low := strings.ToLower(line)
			if strings.Contains(low, "vga") || strings.Contains(low, "3d") || strings.Contains(low, "display") {
				parts := strings.SplitN(line, ": ", 2)
				name := line
				if len(parts) == 2 {
					name = parts[1]
				}
				gpus = append(gpus, GPU{Name: name, VRAM: "—", Driver: "—", Status: "—"})
			}
		}
		return gpus

	case "darwin":
		raw := run("system_profiler", "SPDisplaysDataType")
		chipset := field(raw, "Chipset Model")
		vram := field(raw, "VRAM")
		if chipset == "" {
			chipset = "See: system_profiler SPDisplaysDataType"
		}
		return []GPU{{Name: chipset, VRAM: vram, Driver: "—", Status: "—"}}
	}
	return nil
}

func scanSystem() SysInfo {
	switch runtime.GOOS {
	case "linux":
		return SysInfo{
			Manufacturer: run("cat", "/sys/class/dmi/id/sys_vendor"),
			Model:        run("cat", "/sys/class/dmi/id/product_name"),
			BIOSVersion:  run("cat", "/sys/class/dmi/id/bios_version"),
			BIOSVendor:   run("cat", "/sys/class/dmi/id/bios_vendor"),
			Serial:       run("cat", "/sys/class/dmi/id/product_serial"),
		}
	case "darwin":
		hw := run("system_profiler", "SPHardwareDataType")
		return SysInfo{
			Model:       field(hw, "Model Name"),
			Serial:      field(hw, "Serial Number"),
			BIOSVersion: field(hw, "Boot ROM Version"),
		}
	}
	return SysInfo{}
}
