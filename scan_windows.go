//go:build windows

package main

import (
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

func enableAnsiColors() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	handle, _ := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	var mode uint32
	getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	setConsoleMode.Call(uintptr(handle), uintptr(mode|0x0004))
}

// psExe returns the full path to powershell.exe — avoids relying on PATH
// which may be incomplete when launching a .exe by double-click.
func psExe() string {
	candidates := []string{
		`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`,
		`C:\Windows\SysWOW64\WindowsPowerShell\v1.0\powershell.exe`,
		"powershell.exe",
	}
	for _, p := range candidates {
		if _, err := exec.LookPath(p); err == nil {
			return p
		}
	}
	return "powershell.exe"
}

// psh pipes the script via stdin using -Command - which is the most reliable
// way to pass multiline scripts without temp files or escaping issues.
func psh(script string) string {
	cmd := exec.Command(
		psExe(),
		"-NoProfile", "-NonInteractive",
		"-ExecutionPolicy", "Bypass",
		"-Command", "-",
	)
	cmd.Stdin = strings.NewReader(script)
	out, _ := cmd.Output()
	return strings.TrimSpace(strings.ReplaceAll(string(out), "\r", ""))
}

func parseKV(raw string) map[string]string {
	m := make(map[string]string)
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if idx := strings.Index(line, "="); idx > 0 {
			m[line[:idx]] = strings.TrimSpace(line[idx+1:])
		}
	}
	return m
}

func scanCPU() CPUResult {
	raw := psh(`
$cpu = Get-CimInstance Win32_Processor
$reg = (Get-ItemProperty 'HKLM:\HARDWARE\DESCRIPTION\System\CentralProcessor\0').ProcessorNameString
"REG=" + $reg.Trim()
"WMI=" + $cpu.Name.Trim()
"CORES=" + $cpu.NumberOfCores
"THREADS=" + $cpu.NumberOfLogicalProcessors
"SPEED=" + $cpu.MaxClockSpeed
"GHZ=" + [math]::Round($cpu.MaxClockSpeed / 1000, 2)
"ARCH=" + (switch ($cpu.Architecture) { 0 { "x86" } 9 { "x64 (AMD64)" } 12 { "ARM64" } default { "Unknown" } })
"L2=" + [math]::Round($cpu.L2CacheSize / 1024, 1)
"L3=" + [math]::Round($cpu.L3CacheSize / 1024, 1)
`)
	kv := parseKV(raw)
	return CPUResult{
		RegistryName: kv["REG"],
		WMIName:      kv["WMI"],
		Cores:        kv["CORES"],
		Threads:      kv["THREADS"],
		SpeedMHz:     kv["SPEED"],
		SpeedGHz:     kv["GHZ"],
		Arch:         kv["ARCH"],
		L2:           kv["L2"],
		L3:           kv["L3"],
	}
}

func scanRAM() ([]RAMSlot, string) {
	totalRaw := psh(`[math]::Round((Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory / 1GB, 2)`)

	raw := psh(`
$i = 1
foreach ($s in (Get-CimInstance Win32_PhysicalMemory)) {
    $type = switch ($s.SMBIOSMemoryType) {
        20 { "DDR" } 21 { "DDR2" } 24 { "DDR3" } 26 { "DDR4" } 34 { "DDR5" } default { "Unknown" }
    }
    $mfr  = if ($s.Manufacturer) { $s.Manufacturer.Trim() } else { "N/A" }
    $part = if ($s.PartNumber)   { $s.PartNumber.Trim()   } else { "N/A" }
    "SLOT=" + $i + "|" + [math]::Round($s.Capacity / 1GB, 2) + "|" + $type + "|" + $s.Speed + "|" + $mfr + "|" + $part
    $i++
}
`)
	var slots []RAMSlot
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "SLOT=") {
			continue
		}
		parts := strings.SplitN(strings.TrimPrefix(line, "SLOT="), "|", 6)
		if len(parts) < 6 {
			continue
		}
		slots = append(slots, RAMSlot{
			Slot:         parts[0],
			CapacityGB:   parts[1],
			Type:         parts[2],
			SpeedMHz:     parts[3],
			Manufacturer: parts[4],
			PartNumber:   parts[5],
		})
	}
	return slots, totalRaw
}

func scanStorage() []Disk {
	raw := psh(`
foreach ($d in (Get-CimInstance Win32_DiskDrive)) {
    $size   = if ($d.Size)         { [math]::Round($d.Size / 1GB, 1) } else { "N/A" }
    $serial = if ($d.SerialNumber) { $d.SerialNumber.Trim()           } else { "N/A" }
    "DISK=" + $d.Caption + "|" + $size + "|" + $d.InterfaceType + "|" + $serial
}
`)
	var disks []Disk
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "DISK=") {
			continue
		}
		parts := strings.SplitN(strings.TrimPrefix(line, "DISK="), "|", 4)
		if len(parts) < 4 {
			continue
		}
		disks = append(disks, Disk{
			Caption:   parts[0],
			SizeGB:    parts[1],
			Interface: parts[2],
			Serial:    parts[3],
		})
	}
	return disks
}

func scanGPU() []GPU {
	raw := psh(`
foreach ($g in (Get-CimInstance Win32_VideoController)) {
    $vram = if ($g.AdapterRAM -and $g.AdapterRAM -gt 0) {
        [math]::Round($g.AdapterRAM / 1MB, 0).ToString() + " MB"
    } else { "Shared / N/A" }
    "GPU=" + $g.Caption + "|" + $vram + "|" + $g.DriverVersion + "|" + $g.Status
}
`)
	var gpus []GPU
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "GPU=") {
			continue
		}
		parts := strings.SplitN(strings.TrimPrefix(line, "GPU="), "|", 4)
		if len(parts) < 4 {
			continue
		}
		gpus = append(gpus, GPU{
			Name:   parts[0],
			VRAM:   parts[1],
			Driver: parts[2],
			Status: parts[3],
		})
	}
	return gpus
}

func scanSystem() SysInfo {
	raw := psh(`
$cs  = Get-CimInstance Win32_ComputerSystem
$bio = Get-CimInstance Win32_BIOS
"MFR="     + $cs.Manufacturer
"MODEL="   + $cs.Model
"BIOSVER=" + $bio.SMBIOSBIOSVersion
"BIOSMFR=" + $bio.Manufacturer
"SERIAL="  + $bio.SerialNumber
`)
	kv := parseKV(raw)
	return SysInfo{
		Manufacturer: kv["MFR"],
		Model:        kv["MODEL"],
		BIOSVersion:  kv["BIOSVER"],
		BIOSVendor:   kv["BIOSMFR"],
		Serial:       kv["SERIAL"],
	}
}
