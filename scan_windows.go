//go:build windows

package main

import (
	"encoding/base64"
	"os/exec"
	"strings"
	"syscall"
	"unicode/utf16"
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
	for _, p := range []string{
		`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`,
		`C:\Windows\SysWOW64\WindowsPowerShell\v1.0\powershell.exe`,
		"powershell.exe",
	} {
		if _, err := exec.LookPath(p); err == nil {
			return p
		}
	}
	return "powershell.exe"
}

// psh encodes the script as UTF-16LE + base64 and passes it via -EncodedCommand.
// This is the most reliable way to hand multiline scripts to PowerShell from Go —
// no temp files, no stdin buffering, no escaping issues.
func psh(script string) string {
	// prepend UTF-8 output directive so Go receives clean UTF-8 bytes
	full := "$OutputEncoding=[System.Text.Encoding]::UTF8\n[Console]::OutputEncoding=[System.Text.Encoding]::UTF8\n$ErrorActionPreference='SilentlyContinue'\n" + script

	runes := utf16.Encode([]rune(full))
	b := make([]byte, len(runes)*2)
	for i, r := range runes {
		b[i*2] = byte(r)
		b[i*2+1] = byte(r >> 8)
	}
	encoded := base64.StdEncoding.EncodeToString(b)

	out, _ := exec.Command(
		psExe(),
		"-NoProfile", "-NonInteractive",
		"-ExecutionPolicy", "Bypass",
		"-EncodedCommand", encoded,
	).Output()
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
$a   = $cpu.Architecture
$arch = if ($a -eq 0) { "x86" } elseif ($a -eq 9) { "x64 (AMD64)" } elseif ($a -eq 12) { "ARM64" } else { "Unknown" }
Write-Output ("REG="     + $reg.Trim())
Write-Output ("WMI="     + $cpu.Name.Trim())
Write-Output ("CORES="   + $cpu.NumberOfCores)
Write-Output ("THREADS=" + $cpu.NumberOfLogicalProcessors)
Write-Output ("SPEED="   + $cpu.MaxClockSpeed)
Write-Output ("GHZ="     + [math]::Round($cpu.MaxClockSpeed / 1000, 2))
Write-Output ("ARCH="    + $arch)
Write-Output ("L2="      + [math]::Round($cpu.L2CacheSize / 1024, 1))
Write-Output ("L3="      + [math]::Round($cpu.L3CacheSize / 1024, 1))
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
	totalRaw := psh(`Write-Output ([math]::Round((Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory / 1GB, 2))`)

	raw := psh(`
$i = 1
foreach ($s in (Get-CimInstance Win32_PhysicalMemory)) {
    $t = $s.SMBIOSMemoryType
    $type = if ($t -eq 20) { "DDR" } elseif ($t -eq 21) { "DDR2" } elseif ($t -eq 24) { "DDR3" } elseif ($t -eq 26) { "DDR4" } elseif ($t -eq 34) { "DDR5" } else { "Unknown" }
    $mfr  = if ($s.Manufacturer) { $s.Manufacturer.Trim() } else { "N/A" }
    $part = if ($s.PartNumber)   { $s.PartNumber.Trim()   } else { "N/A" }
    $cap  = [math]::Round($s.Capacity / 1GB, 2)
    Write-Output ("SLOT=" + $i + "|" + $cap + "|" + $type + "|" + $s.Speed + "|" + $mfr + "|" + $part)
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
    $size = "N/A"
    if ($d.Size -gt 0) { $size = [math]::Round($d.Size / 1GB, 1) }
    $serial = if ($d.SerialNumber) { $d.SerialNumber.Trim() } else { "N/A" }
    $iface  = if ($d.InterfaceType) { $d.InterfaceType } else { "N/A" }
    Write-Output ("DISK=" + $d.Caption + "|" + $size + "|" + $iface + "|" + $serial)
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
			Size:      parts[1] + " GB",
			Interface: parts[2],
			Serial:    parts[3],
		})
	}
	return disks
}

func scanGPU() []GPU {
	raw := psh(`
foreach ($g in (Get-CimInstance Win32_VideoController)) {
    $vram = "Shared / N/A"
    if ($g.AdapterRAM -and $g.AdapterRAM -gt 0) {
        $vram = [math]::Round($g.AdapterRAM / 1MB, 0).ToString() + " MB"
    }
    $driver = if ($g.DriverVersion) { $g.DriverVersion } else { "N/A" }
    $status = if ($g.Status)        { $g.Status }        else { "N/A" }
    Write-Output ("GPU=" + $g.Caption + "|" + $vram + "|" + $driver + "|" + $status)
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
Write-Output ("MFR="     + $cs.Manufacturer)
Write-Output ("MODEL="   + $cs.Model)
Write-Output ("BIOSVER=" + $bio.SMBIOSBIOSVersion)
Write-Output ("BIOSMFR=" + $bio.Manufacturer)
Write-Output ("SERIAL="  + $bio.SerialNumber)
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
