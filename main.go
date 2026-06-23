package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ── Colors ────────────────────────────────────────────────────────────────────
const (
	cRed     = "\033[0;31m"
	cGreen   = "\033[0;32m"
	cYellow  = "\033[1;33m"
	cCyan    = "\033[0;36m"
	cMagenta = "\033[0;35m"
	cBold    = "\033[1m"
	cDim     = "\033[2m"
	cReset   = "\033[0m"
)

// ── Spinner ───────────────────────────────────────────────────────────────────
type spinner struct {
	done chan struct{}
	wg   sync.WaitGroup
}

func newSpinner(msg string) *spinner {
	s := &spinner{done: make(chan struct{})}
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		i := 0
		for {
			select {
			case <-s.done:
				fmt.Print("\r\033[K")
				return
			default:
				fmt.Printf("\r  %s%s%s  %s", cCyan, frames[i%len(frames)], cReset, msg)
				time.Sleep(80 * time.Millisecond)
				i++
			}
		}
	}()
	return s
}

func (s *spinner) stop() {
	close(s.done)
	s.wg.Wait()
}

// ── UI helpers ────────────────────────────────────────────────────────────────
func typewrite(text string, delay time.Duration) {
	for _, ch := range text {
		fmt.Printf("%c", ch)
		time.Sleep(delay)
	}
	fmt.Println()
}

func reveal(line string) {
	time.Sleep(55 * time.Millisecond)
	fmt.Println(line)
}

func section(title string) {
	fmt.Println()
	time.Sleep(250 * time.Millisecond)
	fmt.Printf("%s%s▶ %s%s\n", cCyan, cBold, title, cReset)
	fmt.Printf("%s%s%s\n", cCyan, strings.Repeat("─", 56), cReset)
}

// ── ASCII art ─────────────────────────────────────────────────────────────────
func drawArt() {
	type artLine struct{ text, color string }
	lines := []artLine{
		{"  ██████╗ ███████╗██╗   ██╗██╗ ██████╗███████╗", cCyan + cBold},
		{"  ██╔══██╗██╔════╝██║   ██║██║██╔════╝██╔════╝", cCyan + cBold},
		{"  ██║  ██║█████╗  ╚██╗ ██╔╝██║██║     █████╗  ", cCyan + cBold},
		{"  ██║  ██║██╔══╝   ╚████╔╝ ██║██║     ██╔══╝  ", cCyan + cBold},
		{"  ██████╔╝███████╗  ╚██╔╝  ██║╚██████╗███████╗", cCyan + cBold},
		{"  ╚═════╝ ╚══════╝   ╚═╝   ╚═╝ ╚═════╝╚══════╝", cCyan + cBold},
		{"", ""},
		{"   ███████╗███████╗███╗   ██╗███████╗███████╗", cMagenta + cBold},
		{"   ██╔════╝██╔════╝████╗  ██║██╔════╝██╔════╝", cMagenta + cBold},
		{"   ███████╗█████╗  ██╔██╗ ██║███████╗█████╗  ", cMagenta + cBold},
		{"   ╚════██║██╔══╝  ██║╚██╗██║╚════██║██╔══╝  ", cMagenta + cBold},
		{"   ███████║███████╗██║ ╚████║███████║███████╗", cMagenta + cBold},
		{"   ╚══════╝╚══════╝╚═╝  ╚═══╝╚══════╝╚══════╝", cMagenta + cBold},
	}
	fmt.Println()
	for _, l := range lines {
		if l.text == "" {
			fmt.Println()
			time.Sleep(120 * time.Millisecond)
			continue
		}
		fmt.Printf("%s%s%s\n", l.color, l.text, cReset)
		time.Sleep(50 * time.Millisecond)
	}
	fmt.Println()
}

// ── Support banner ────────────────────────────────────────────────────────────
// Inner box width = 54 visual chars. Total line = 2 indent + 1 ║ + 54 + 1 ║ = 58
func drawSupportBanner() {
	fmt.Println()
	time.Sleep(300 * time.Millisecond)

	b := cMagenta + cBold
	m := cMagenta
	r := cReset
	g := cGreen
	c := cCyan
	y := cYellow

	banner := []string{
		fmt.Sprintf("%s  ╔══════════════════════════════════════════════════════╗%s", b, r),
		fmt.Sprintf("%s  ║            ★  SUPPORT JEFFREYON  ★                  ║%s", b, r),
		fmt.Sprintf("%s  ╠══════════════════════════════════════════════════════╣%s", m, r),
		fmt.Sprintf("%s  ║  This tool is free — sharing it is how it grows.    ║%s", m, r),
		fmt.Sprintf("%s  ╠══════════════════════════════════════════════════════╣%s", m, r),
		fmt.Sprintf("  %s║%s  %s→%s Help your friend scan their computer              %s║%s", m, r, g, r, m, r),
		fmt.Sprintf("  %s║%s    Share this script — might just save them money    %s║%s", m, r, m, r),
		fmt.Sprintf("  %s║%s                                                      %s║%s", m, r, m, r),
		fmt.Sprintf("  %s║%s  %s→%s Refer me for a project                            %s║%s", m, r, g, r, m, r),
		fmt.Sprintf("  %s║%s    %shttps://wa.link/b11q29%s                            %s║%s", m, r, c, r, m, r),
		fmt.Sprintf("  %s║%s                                                      %s║%s", m, r, m, r),
		fmt.Sprintf("  %s║%s  %s→%s Send money for data and pizza                     %s║%s", m, r, g, r, m, r),
		fmt.Sprintf("  %s║%s    %s8085709543 — Opay%s                                %s║%s", m, r, y, r, m, r),
		fmt.Sprintf("  %s║%s    (anything your hand reach)                        %s║%s", m, r, m, r),
		fmt.Sprintf("%s  ╚══════════════════════════════════════════════════════╝%s", b, r),
	}
	for _, line := range banner {
		time.Sleep(50 * time.Millisecond)
		fmt.Println(line)
	}
	fmt.Println()
}

// ── Main ──────────────────────────────────────────────────────────────────────
func main() {
	enableAnsiColors()
	fmt.Print("\033[H\033[2J") // clear screen

	drawArt()

	fmt.Printf("  %s", cDim)
	typewrite("Hardware Verifier — reads from chip, not from the label", 18*time.Millisecond)
	fmt.Printf("  ")
	typewrite("Detects registry tricks used to misrepresent laptop specs", 18*time.Millisecond)
	fmt.Print(cReset)
	time.Sleep(500 * time.Millisecond)

	mismatch := false

	// ── CPU ──────────────────────────────────────────────────────────────────
	section("CPU VERIFICATION")

	sp := newSpinner("Reading registry label (what Windows shows you)...")
	// trigger the registry read concurrently with the spinner
	type cpuResult struct{ r CPUResult }
	cpuCh := make(chan cpuResult, 1)
	go func() { cpuCh <- cpuResult{scanCPU()} }()
	cpu := (<-cpuCh).r
	sp.stop()

	reveal(fmt.Sprintf("  Shown in Settings : %s%s%s", cYellow, cpu.RegistryName, cReset))
	reveal(fmt.Sprintf("  Hardware (WMI)    : %s%s%s", cGreen, cpu.WMIName, cReset))
	fmt.Println()

	if cpu.RegistryName == cpu.WMIName {
		reveal(fmt.Sprintf("  Status            : %s✓  Match — no tampering detected%s", cGreen, cReset))
	} else {
		reveal(fmt.Sprintf("  Status            : %s⚠  MISMATCH! Description has been altered!%s", cRed, cReset))
		reveal(fmt.Sprintf("  %s     Real CPU differs from what Windows Settings shows.%s", cRed, cReset))
		mismatch = true
	}

	fmt.Println()
	reveal(fmt.Sprintf("  Physical Cores    : %s%s%s", cBold, cpu.Cores, cReset))
	reveal(fmt.Sprintf("  Logical Threads   : %s%s%s", cBold, cpu.Threads, cReset))
	reveal(fmt.Sprintf("  Max Clock Speed   : %s%s MHz  (~%s GHz)%s", cBold, cpu.SpeedMHz, cpu.SpeedGHz, cReset))
	reveal(fmt.Sprintf("  Architecture      : %s%s%s", cBold, cpu.Arch, cReset))
	reveal(fmt.Sprintf("  L2 Cache          : %s%s MB%s", cBold, cpu.L2, cReset))
	reveal(fmt.Sprintf("  L3 Cache          : %s%s MB%s", cBold, cpu.L3, cReset))

	// ── RAM ──────────────────────────────────────────────────────────────────
	section("RAM VERIFICATION  (reads from SMBIOS firmware)")

	sp = newSpinner("Enumerating physical memory slots...")
	slots, totalRAM := scanRAM()
	sp.stop()

	reveal(fmt.Sprintf("  Total Physical RAM : %s%s GB%s", cBold, totalRAM, cReset))
	fmt.Println()
	for _, s := range slots {
		reveal(fmt.Sprintf("  Slot %s : %s GB | %s @ %s MHz | Mfr: %s | Part: %s",
			s.Slot, s.CapacityGB, s.Type, s.SpeedMHz, s.Manufacturer, s.PartNumber))
	}

	// ── Storage ───────────────────────────────────────────────────────────────
	section("STORAGE VERIFICATION")

	sp = newSpinner("Reading drive identifiers from device drivers...")
	disks := scanStorage()
	sp.stop()

	for _, d := range disks {
		reveal(fmt.Sprintf("  Drive  : %s", d.Caption))
		reveal(fmt.Sprintf("  Size   : %s GB  |  Interface: %s  |  Serial: %s", d.SizeGB, d.Interface, d.Serial))
		fmt.Println()
	}

	// ── GPU ───────────────────────────────────────────────────────────────────
	section("GPU VERIFICATION")

	sp = newSpinner("Polling video controller...")
	gpus := scanGPU()
	sp.stop()

	for _, g := range gpus {
		reveal(fmt.Sprintf("  GPU    : %s", g.Name))
		reveal(fmt.Sprintf("  VRAM   : %s  |  Driver: %s  |  Status: %s", g.VRAM, g.Driver, g.Status))
		fmt.Println()
	}

	// ── System identity ───────────────────────────────────────────────────────
	section("SYSTEM IDENTITY")

	sp = newSpinner("Reading BIOS and machine identity...")
	sys := scanSystem()
	sp.stop()

	reveal(fmt.Sprintf("  Manufacturer : %s", sys.Manufacturer))
	reveal(fmt.Sprintf("  Model        : %s", sys.Model))
	reveal(fmt.Sprintf("  BIOS Version : %s", sys.BIOSVersion))
	reveal(fmt.Sprintf("  BIOS Vendor  : %s", sys.BIOSVendor))
	reveal(fmt.Sprintf("  Serial No.   : %s", sys.Serial))

	// ── Verdict ───────────────────────────────────────────────────────────────
	fmt.Println()
	time.Sleep(400 * time.Millisecond)

	if mismatch {
		for _, line := range []string{
			fmt.Sprintf("%s%s  ╔══════════════════════════════════════════════════════╗%s", cRed, cBold, cReset),
			fmt.Sprintf("%s%s  ║  ⚠  WARNING: Spec mismatch detected!                ║%s", cRed, cBold, cReset),
			fmt.Sprintf("%s%s  ║  The CPU name shown in Windows Settings was altered. ║%s", cRed, cBold, cReset),
			fmt.Sprintf("%s%s  ║  Do NOT pay for specs you cannot verify above.       ║%s", cRed, cBold, cReset),
			fmt.Sprintf("%s%s  ╚══════════════════════════════════════════════════════╝%s", cRed, cBold, cReset),
		} {
			time.Sleep(60 * time.Millisecond)
			fmt.Println(line)
		}
	} else {
		for _, line := range []string{
			fmt.Sprintf("%s%s  ╔══════════════════════════════════════════════════════╗%s", cGreen, cBold, cReset),
			fmt.Sprintf("%s%s  ║  ✓  All checks passed — specs appear consistent.     ║%s", cGreen, cBold, cReset),
			fmt.Sprintf("%s%s  ║  Hardware reads match displayed descriptions.        ║%s", cGreen, cBold, cReset),
			fmt.Sprintf("%s%s  ╚══════════════════════════════════════════════════════╝%s", cGreen, cBold, cReset),
		} {
			time.Sleep(60 * time.Millisecond)
			fmt.Println(line)
		}
	}

	fmt.Println()
	fmt.Printf("  %sHKLM\\HARDWARE\\DESCRIPTION\\System\\CentralProcessor\\0%s\n", cDim, cReset)
	fmt.Printf("  %s└─ just a display string, editable by any admin.%s\n", cDim, cReset)
	fmt.Printf("  %s   WMI reads above query CPUID directly from the chip.%s\n", cDim, cReset)

	drawSupportBanner()
}
