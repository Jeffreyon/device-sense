#!/bin/bash
# device-sense.sh — Animated hardware verifier
# Reads from WMI/SMBIOS (hardware level) and cross-checks the registry
# display string that sellers can edit to fake CPU specs in Windows Settings.

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

MISMATCH_FOUND=0
SPIN_PID=""

# ── Helpers ────────────────────────────────────────────────────────────────────
psh() { powershell.exe -NoProfile -NonInteractive -Command "$1" 2>/dev/null | tr -d '\r'; }

typewrite() {
    local text="$1" delay="${2:-0.022}"
    for ((i=0; i<${#text}; i++)); do
        printf "%s" "${text:$i:1}"; sleep "$delay"
    done
    echo ""
}

start_spinner() {
    local msg="$1"
    (
        local f=('⠋' '⠙' '⠹' '⠸' '⠼' '⠴' '⠦' '⠧' '⠇' '⠏')
        while true; do
            for frame in "${f[@]}"; do
                printf "\r  \033[0;36m%s\033[0m  %s" "$frame" "$msg"
                sleep 0.08
            done
        done
    ) &
    SPIN_PID=$!
}

stop_spinner() {
    [ -z "$SPIN_PID" ] && return
    { kill "$SPIN_PID"; wait "$SPIN_PID"; } 2>/dev/null
    printf "\r\033[K"
    SPIN_PID=""
}

reveal() { sleep "${2:-0.055}"; echo -e "$1"; }

section() {
    echo ""
    sleep 0.25
    echo -e "${CYAN}${BOLD}▶ $1${NC}"
    echo -e "${CYAN}$(printf '─%.0s' {1..56})${NC}"
}

# ══════════════════════════════════════════════════════════════════════════════
# BOOT — ASCII art drawn line by line
# ══════════════════════════════════════════════════════════════════════════════
clear
echo ""
echo -e "${CYAN}${BOLD}"
sleep 0.05; echo '  ██████╗ ███████╗██╗   ██╗██╗ ██████╗███████╗'
sleep 0.05; echo '  ██╔══██╗██╔════╝██║   ██║██║██╔════╝██╔════╝'
sleep 0.05; echo '  ██║  ██║█████╗  ╚██╗ ██╔╝██║██║     █████╗  '
sleep 0.05; echo '  ██║  ██║██╔══╝   ╚████╔╝ ██║██║     ██╔══╝  '
sleep 0.05; echo '  ██████╔╝███████╗  ╚██╔╝  ██║╚██████╗███████╗'
sleep 0.05; echo '  ╚═════╝ ╚══════╝   ╚═╝   ╚═╝ ╚═════╝╚══════╝'
echo -e "${NC}"
sleep 0.12
echo -e "${MAGENTA}${BOLD}"
sleep 0.05; echo '   ███████╗███████╗███╗   ██╗███████╗███████╗'
sleep 0.05; echo '   ██╔════╝██╔════╝████╗  ██║██╔════╝██╔════╝'
sleep 0.05; echo '   ███████╗█████╗  ██╔██╗ ██║███████╗█████╗  '
sleep 0.05; echo '   ╚════██║██╔══╝  ██║╚██╗██║╚════██║██╔══╝  '
sleep 0.05; echo '   ███████║███████╗██║ ╚████║███████║███████╗'
sleep 0.05; echo '   ╚══════╝╚══════╝╚═╝  ╚═══╝╚══════╝╚══════╝'
echo -e "${NC}"
sleep 0.2

printf "  ${DIM}"
typewrite "Hardware Verifier — reads from chip, not from the label" 0.018
printf "  ${DIM}"
typewrite "Detects registry tricks used to misrepresent laptop specs" 0.018
echo -e "${NC}"
sleep 0.5

# ══════════════════════════════════════════════════════════════════════════════
# CPU
# ══════════════════════════════════════════════════════════════════════════════
section "CPU VERIFICATION"

start_spinner "Reading registry label (what Windows shows you)..."
REGISTRY_CPU=$(psh "(Get-ItemProperty 'HKLM:\HARDWARE\DESCRIPTION\System\CentralProcessor\0').ProcessorNameString")
stop_spinner

start_spinner "Querying WMI — pulling CPUID direct from hardware..."
WMI_CPU=$(psh "(Get-WmiObject Win32_Processor).Name.Trim()")
CORES=$(psh "(Get-WmiObject Win32_Processor).NumberOfCores")
THREADS=$(psh "(Get-WmiObject Win32_Processor).NumberOfLogicalProcessors")
SPEED_MHZ=$(psh "(Get-WmiObject Win32_Processor).MaxClockSpeed")
SPEED_GHZ=$(psh "[math]::Round($SPEED_MHZ/1000,2)")
ARCH=$(psh "switch((Get-WmiObject Win32_Processor).Architecture){0{'x86'} 9{'x64 (AMD64)'} 12{'ARM64'} default{'Unknown'}}")
L2=$(psh "[math]::Round((Get-WmiObject Win32_Processor).L2CacheSize / 1024, 1)")
L3=$(psh "[math]::Round((Get-WmiObject Win32_Processor).L3CacheSize / 1024, 1)")
stop_spinner

reveal "  Shown in Settings : ${YELLOW}${REGISTRY_CPU}${NC}"
reveal "  Hardware (WMI)    : ${GREEN}${WMI_CPU}${NC}"
echo ""

if [ "$REGISTRY_CPU" = "$WMI_CPU" ]; then
    reveal "  Status            : ${GREEN}✓  Match — no tampering detected${NC}"
else
    reveal "  Status            : ${RED}⚠  MISMATCH! Description has been altered!${NC}"
    reveal "  ${RED}     Real CPU differs from what Windows Settings displays.${NC}"
    MISMATCH_FOUND=1
fi

echo ""
reveal "  Physical Cores    : ${BOLD}${CORES}${NC}"
reveal "  Logical Threads   : ${BOLD}${THREADS}${NC}"
reveal "  Max Clock Speed   : ${BOLD}${SPEED_MHZ} MHz  (~${SPEED_GHZ} GHz)${NC}"
reveal "  Architecture      : ${BOLD}${ARCH}${NC}"
reveal "  L2 Cache          : ${BOLD}${L2} MB${NC}"
reveal "  L3 Cache          : ${BOLD}${L3} MB${NC}"

if echo "$WMI_CPU" | grep -qi "i3" && [ "$CORES" -gt 4 ]; then
    reveal "  ${YELLOW}⚠  Core count unusually high for an i3 — verify generation${NC}"
fi
if echo "$WMI_CPU" | grep -qi "i5" && [ -n "$CORES" ] && [ "$CORES" -lt 4 ]; then
    reveal "  ${YELLOW}⚠  Core count unusually low for an i5 — could be 7th-gen or older${NC}"
fi
if echo "$WMI_CPU" | grep -qi "i9" && [ -n "$CORES" ] && [ "$CORES" -lt 8 ]; then
    reveal "  ${YELLOW}⚠  Core count unusually low for an i9 — possible misrepresentation${NC}"
fi

# ══════════════════════════════════════════════════════════════════════════════
# RAM
# ══════════════════════════════════════════════════════════════════════════════
section "RAM VERIFICATION  (reads from SMBIOS firmware)"

start_spinner "Enumerating physical memory slots..."
TOTAL_RAM=$(psh "[math]::Round((Get-WmiObject Win32_ComputerSystem).TotalPhysicalMemory / 1GB, 2)")
RAM_DETAILS=$(psh "
\$i = 1
foreach (\$s in (Get-WmiObject Win32_PhysicalMemory)) {
    \$cap  = [math]::Round(\$s.Capacity / 1GB, 2)
    \$type = switch(\$s.SMBIOSMemoryType) {
        20{'DDR'} 21{'DDR2'} 24{'DDR3'} 26{'DDR4'} 34{'DDR5'} default{'Unknown'}
    }
    \$mfr  = if(\$s.Manufacturer -and \$s.Manufacturer.Trim() -ne '') { \$s.Manufacturer.Trim() } else { 'N/A' }
    \$part = if(\$s.PartNumber) { \$s.PartNumber.Trim() } else { 'N/A' }
    Write-Host \"  Slot \$i : \$cap GB | \$type @ \$(\$s.Speed) MHz | Mfr: \$mfr | Part: \$part\"
    \$i++
}
")
stop_spinner

reveal "  Total Physical RAM : ${BOLD}${TOTAL_RAM} GB${NC}"
echo ""
while IFS= read -r line; do
    [ -n "$line" ] && reveal "$line"
done <<< "$RAM_DETAILS"

# ══════════════════════════════════════════════════════════════════════════════
# STORAGE
# ══════════════════════════════════════════════════════════════════════════════
section "STORAGE VERIFICATION"

start_spinner "Reading drive identifiers from device drivers..."
STORAGE_DETAILS=$(psh "
foreach (\$d in (Get-WmiObject Win32_DiskDrive)) {
    \$size   = if(\$d.Size)        { [math]::Round(\$d.Size / 1GB, 1).ToString() + ' GB' } else { 'N/A' }
    \$serial = if(\$d.SerialNumber) { \$d.SerialNumber.Trim() } else { 'N/A' }
    Write-Host \"  Drive  : \$(\$d.Caption)\"
    Write-Host \"  Size   : \$size  |  Interface: \$(\$d.InterfaceType)  |  Serial: \$serial\"
    Write-Host ''
}
")
stop_spinner

while IFS= read -r line; do
    reveal "$line" 0.05
done <<< "$STORAGE_DETAILS"

# ══════════════════════════════════════════════════════════════════════════════
# GPU
# ══════════════════════════════════════════════════════════════════════════════
section "GPU VERIFICATION"

start_spinner "Polling video controller..."
GPU_DETAILS=$(psh "
foreach (\$g in (Get-WmiObject Win32_VideoController)) {
    \$vram = if(\$g.AdapterRAM -and \$g.AdapterRAM -gt 0) {
        [math]::Round(\$g.AdapterRAM / 1MB, 0).ToString() + ' MB'
    } else { 'Shared / N/A' }
    Write-Host \"  GPU    : \$(\$g.Caption)\"
    Write-Host \"  VRAM   : \$vram  |  Driver: \$(\$g.DriverVersion)  |  Status: \$(\$g.Status)\"
    Write-Host ''
}
")
stop_spinner

while IFS= read -r line; do
    reveal "$line" 0.05
done <<< "$GPU_DETAILS"

# ══════════════════════════════════════════════════════════════════════════════
# SYSTEM IDENTITY
# ══════════════════════════════════════════════════════════════════════════════
section "SYSTEM IDENTITY"

start_spinner "Reading BIOS and machine identity..."
IDENTITY=$(psh "
\$cs  = Get-WmiObject Win32_ComputerSystem
\$bio = Get-WmiObject Win32_BIOS
Write-Host \"  Manufacturer : \$(\$cs.Manufacturer)\"
Write-Host \"  Model        : \$(\$cs.Model)\"
Write-Host \"  BIOS Version : \$(\$bio.SMBIOSBIOSVersion)\"
Write-Host \"  BIOS Vendor  : \$(\$bio.Manufacturer)\"
Write-Host \"  Serial No.   : \$(\$bio.SerialNumber)\"
")
stop_spinner

while IFS= read -r line; do
    reveal "$line"
done <<< "$IDENTITY"

# ══════════════════════════════════════════════════════════════════════════════
# VERDICT
# ══════════════════════════════════════════════════════════════════════════════
echo ""
sleep 0.4

if [ "$MISMATCH_FOUND" -eq 1 ]; then
    sleep 0.05; echo -e "${BOLD}${RED}  ╔══════════════════════════════════════════════════════╗${NC}"
    sleep 0.06; echo -e "${BOLD}${RED}  ║  ⚠  WARNING: Spec mismatch detected!                ║${NC}"
    sleep 0.06; echo -e "${BOLD}${RED}  ║  The CPU name shown in Windows Settings was altered. ║${NC}"
    sleep 0.06; echo -e "${BOLD}${RED}  ║  Do NOT pay for specs you cannot verify above.       ║${NC}"
    sleep 0.05; echo -e "${BOLD}${RED}  ╚══════════════════════════════════════════════════════╝${NC}"
else
    sleep 0.05; echo -e "${BOLD}${GREEN}  ╔══════════════════════════════════════════════════════╗${NC}"
    sleep 0.06; echo -e "${BOLD}${GREEN}  ║  ✓  All checks passed — specs appear consistent.     ║${NC}"
    sleep 0.06; echo -e "${BOLD}${GREEN}  ║  Hardware reads match displayed descriptions.        ║${NC}"
    sleep 0.05; echo -e "${BOLD}${GREEN}  ╚══════════════════════════════════════════════════════╝${NC}"
fi

echo ""
echo -e "  ${DIM}HKLM\\HARDWARE\\DESCRIPTION\\System\\CentralProcessor\\0${NC}"
echo -e "  ${DIM}└─ just a display string, editable by any admin.${NC}"
echo -e "  ${DIM}   WMI reads above query CPUID directly from the chip.${NC}"

# ══════════════════════════════════════════════════════════════════════════════
# SUPPORT JEFFREYON
# Inner box width = 54 chars. Each content line: ║ + [54 chars] + ║
# ══════════════════════════════════════════════════════════════════════════════
echo ""
sleep 0.3

sleep 0.04; echo -e "${MAGENTA}${BOLD}  ╔══════════════════════════════════════════════════════╗${NC}"
sleep 0.05; echo -e "${MAGENTA}${BOLD}  ║            ★  SUPPORT JEFFREYON  ★                  ║${NC}"
sleep 0.04; echo -e "${MAGENTA}  ╠══════════════════════════════════════════════════════╣${NC}"
sleep 0.05; echo -e "${MAGENTA}  ║  This tool is free — sharing it is how it grows.    ║${NC}"
sleep 0.04; echo -e "${MAGENTA}  ╠══════════════════════════════════════════════════════╣${NC}"
sleep 0.08; echo -e "  ${MAGENTA}║${NC}  ${GREEN}→${NC} Help your friend scan their computer              ${MAGENTA}║${NC}"
sleep 0.05; echo -e "  ${MAGENTA}║${NC}    Share this script — might just save them money    ${MAGENTA}║${NC}"
sleep 0.05; echo -e "  ${MAGENTA}║${NC}                                                      ${MAGENTA}║${NC}"
sleep 0.08; echo -e "  ${MAGENTA}║${NC}  ${GREEN}→${NC} Refer me for a project                            ${MAGENTA}║${NC}"
sleep 0.05; echo -e "  ${MAGENTA}║${NC}    ${CYAN}https://wa.link/b11q29${NC}                            ${MAGENTA}║${NC}"
sleep 0.05; echo -e "  ${MAGENTA}║${NC}                                                      ${MAGENTA}║${NC}"
sleep 0.08; echo -e "  ${MAGENTA}║${NC}  ${GREEN}→${NC} Send money for data and pizza                     ${MAGENTA}║${NC}"
sleep 0.05; echo -e "  ${MAGENTA}║${NC}    ${YELLOW}8085709543 — Opay${NC}                                ${MAGENTA}║${NC}"
sleep 0.05; echo -e "  ${MAGENTA}║${NC}    (anything your hand reach)                 ${MAGENTA}║${NC}"
sleep 0.04; echo -e "${MAGENTA}${BOLD}  ╚══════════════════════════════════════════════════════╝${NC}"
echo ""
