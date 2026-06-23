package main

// CPUResult holds both the registry display string and the real WMI hardware read.
type CPUResult struct {
	RegistryName string
	WMIName      string
	Cores        string
	Threads      string
	SpeedMHz     string
	SpeedGHz     string
	Arch         string
	L2           string
	L3           string
}

type RAMSlot struct {
	Slot         string
	CapacityGB   string
	Type         string
	SpeedMHz     string
	Manufacturer string
	PartNumber   string
}

type Disk struct {
	Caption   string
	SizeGB    string
	Interface string
	Serial    string
}

type GPU struct {
	Name    string
	VRAM    string
	Driver  string
	Status  string
}

type SysInfo struct {
	Manufacturer string
	Model        string
	BIOSVersion  string
	BIOSVendor   string
	Serial       string
}
