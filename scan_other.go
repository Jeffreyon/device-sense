//go:build !windows

package main

import (
	"fmt"
	"os"
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

func isWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	lower := strings.ToLower(string(data))
	return strings.Contains(lower, "microsoft") || strings.Contains(lower, "wsl")
}

func formatBytes(bytesStr string) string {
	var b int64
	fmt.Sscanf(bytesStr, "%d", &b)
	if b == 0 {
		return "—"
	}
	switch {
	case b >= 1_099_511_627_776:
		return fmt.Sprintf("%.1f TB", float64(b)/1_099_511_627_776)
	case b >= 1_073_741_824:
		return fmt.Sprintf("%.1f GB", float64(b)/1_073_741_824)
	default:
		return fmt.Sprintf("%.0f MB", float64(b)/1_048_576)
	}
}

func scanCPU() CPUResult {
	var name, cores, threads, speedMHz, arch, l2, l3 string

	switch runtime.GOOS {
	case "linux":
		raw := run("cat", "/proc/cpuinfo")
		name = field(raw, "model name")
		cores = field(raw, "cpu cores")
		threads = field(raw, "siblings")
		speedMHz = field(raw, "cpu mhz")
		arch = run("uname", "-m")

		// lscpu gives proper cache info
		lscpuRaw := run("lscpu")
		l2raw := field(lscpuRaw, "L2 cache")
		l3raw := field(lscpuRaw, "L3 cache")
		// normalise: lscpu reports "256 KiB" or "3 MiB" — convert to MB
		l2 = normCache(l2raw)
		l3 = normCache(l3raw)

	case "darwin":
		name = run("sysctl", "-n", "machdep.cpu.brand_string")
		cores = run("sysctl", "-n", "hw.physicalcpu")
		threads = run("sysctl", "-n", "hw.logicalcpu")
		hz := run("sysctl", "-n", "hw.cpufrequency_max")
		if hz != "" {
			var hzInt int64
			fmt.Sscanf(hz, "%d", &hzInt)
			speedMHz = fmt.Sprintf("%.0f", float64(hzInt)/1_000_000)
		}
		arch = run("uname", "-m")
		l2Bytes := run("sysctl", "-n", "hw.l2cachesize")
		l3Bytes := run("sysctl", "-n", "hw.l3cachesize")
		l2 = formatBytes(l2Bytes)
		l3 = formatBytes(l3Bytes)
	}

	speedGHz := ""
	if speedMHz != "" {
		var mhz float64
		fmt.Sscanf(speedMHz, "%f", &mhz)
		speedGHz = fmt.Sprintf("%.2f", mhz/1000)
	}

	return CPUResult{
		SkipRegistryCheck: true,
		WMIName:           name,
		Cores:             cores,
		Threads:           threads,
		SpeedMHz:          speedMHz,
		SpeedGHz:          speedGHz,
		Arch:              arch,
		L2:                l2,
		L3:                l3,
	}
}

// normCache converts lscpu cache strings like "256 KiB" or "3 MiB" to "X MB"
func normCache(s string) string {
	if s == "" {
		return ""
	}
	s = strings.TrimSpace(s)
	var val float64
	var unit string
	fmt.Sscanf(s, "%f %s", &val, &unit)
	unit = strings.ToLower(unit)
	switch {
	case strings.HasPrefix(unit, "kib") || strings.HasPrefix(unit, "kb"):
		return fmt.Sprintf("%.1f MB", val/1024)
	case strings.HasPrefix(unit, "mib") || strings.HasPrefix(unit, "mb"):
		return fmt.Sprintf("%.1f MB", val)
	case strings.HasPrefix(unit, "gib") || strings.HasPrefix(unit, "gb"):
		return fmt.Sprintf("%.1f GB", val)
	}
	return s // return raw if we can't parse
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
		b := run("sysctl", "-n", "hw.memsize")
		totalGB := ""
		if b != "" {
			var bInt int64
			fmt.Sscanf(b, "%d", &bInt)
			totalGB = fmt.Sprintf("%.2f", float64(bInt)/1_073_741_824)
		}
		return []RAMSlot{}, totalGB
	}
	return []RAMSlot{}, "N/A"
}

func scanStorage() []Disk {
	switch runtime.GOOS {
	case "linux":
		// -b returns sizes in bytes so we can format them ourselves
		raw := run("lsblk", "-d", "-b", "-o", "NAME,SIZE,TYPE,MODEL", "--noheadings")
		var disks []Disk
		for _, line := range strings.Split(raw, "\n") {
			f := strings.Fields(line)
			if len(f) < 3 || f[2] != "disk" {
				continue
			}
			name := f[0]
			size := formatBytes(f[1])
			model := ""
			if len(f) >= 4 {
				model = strings.Join(f[3:], " ")
			}
			caption := name
			if model != "" {
				caption = name + " (" + model + ")"
			}
			disks = append(disks, Disk{
				Caption:   caption,
				Size:      size,
				Interface: "—",
				Serial:    "—",
			})
		}
		return disks

	case "darwin":
		raw := run("diskutil", "info", "-all")
		if raw == "" {
			return []Disk{{Caption: "Run: diskutil list", Size: "—", Interface: "—", Serial: "—"}}
		}
		var disks []Disk
		var cur Disk
		for _, line := range strings.Split(raw, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Device Node:") {
				if cur.Caption != "" {
					disks = append(disks, cur)
				}
				cur = Disk{Caption: strings.TrimSpace(strings.TrimPrefix(line, "Device Node:")), Interface: "—", Serial: "—"}
			} else if strings.HasPrefix(line, "Disk Size:") {
				cur.Size = strings.TrimSpace(strings.SplitN(strings.TrimPrefix(line, "Disk Size:"), "(", 2)[0])
			} else if strings.HasPrefix(line, "Device / Media Name:") {
				name := strings.TrimSpace(strings.TrimPrefix(line, "Device / Media Name:"))
				if name != "" {
					cur.Caption = cur.Caption + " (" + name + ")"
				}
			}
		}
		if cur.Caption != "" {
			disks = append(disks, cur)
		}
		return disks
	}
	return nil
}

func scanGPU() []GPU {
	switch runtime.GOOS {
	case "linux":
		if isWSL() {
			return []GPU{{Name: "GPU not directly accessible from WSL", VRAM: "—", Driver: "—", Status: "Run natively on Windows for full GPU info"}}
		}
		raw := run("lspci")
		if raw == "" {
			return []GPU{{Name: "lspci not available — install pciutils", VRAM: "—", Driver: "—", Status: "—"}}
		}
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
		if len(gpus) == 0 {
			return []GPU{{Name: "No GPU found via lspci", VRAM: "—", Driver: "—", Status: "—"}}
		}
		return gpus

	case "darwin":
		raw := run("system_profiler", "SPDisplaysDataType")
		chipset := field(raw, "Chipset Model")
		vram := field(raw, "VRAM")
		if chipset == "" {
			chipset = "Could not read — try: system_profiler SPDisplaysDataType"
		}
		return []GPU{{Name: chipset, VRAM: vram, Driver: "—", Status: "—"}}
	}
	return nil
}

func scanSystem() SysInfo {
	switch runtime.GOOS {
	case "linux":
		if isWSL() {
			return SysInfo{
				Manufacturer: "WSL (Windows Subsystem for Linux)",
				Model:        "Virtual environment — run on Windows for real hardware identity",
				BIOSVersion:  "—",
				BIOSVendor:   "—",
				Serial:       "—",
			}
		}
		readFile := func(p string) string {
			b, err := os.ReadFile(p)
			if err != nil {
				return ""
			}
			return strings.TrimSpace(string(b))
		}
		return SysInfo{
			Manufacturer: readFile("/sys/class/dmi/id/sys_vendor"),
			Model:        readFile("/sys/class/dmi/id/product_name"),
			BIOSVersion:  readFile("/sys/class/dmi/id/bios_version"),
			BIOSVendor:   readFile("/sys/class/dmi/id/bios_vendor"),
			Serial:       readFile("/sys/class/dmi/id/product_serial"),
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
