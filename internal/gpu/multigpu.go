package gpu

import (
	"crypto/sha256"
	"fmt"
	"syscall"
	"unsafe"
)

type DeviceInfo struct {
	Index        int
	Name         string
	VendorID     uint32
	DeviceID     uint32
	VRAMBytes    uint64
	UUID         [16]byte // stable across process restarts
	PipelineRank int      // -1 = unassigned
}

type Topology struct {
	Devices      []DeviceInfo
	Interconnect string // "shared_mem" | "pcie" | "loopback"
	SingleDevice bool   // true = only 1 device, pipeline disabled
}

// NumDevices returns the total count of available devices.
func (t *Topology) NumDevices() int {
	if t == nil {
		return 0
	}
	return len(t.Devices)
}

// AssignPipelineStages maps N pipeline stages to available devices in a round-robin wrap-around manner.
func (t *Topology) AssignPipelineStages(nStages int) error {
	if t == nil || len(t.Devices) == 0 {
		return fmt.Errorf("topology contains no available compute devices")
	}
	if nStages < 1 {
		return fmt.Errorf("pipeline stages must be at least 1, got %d", nStages)
	}

	for i := range t.Devices {
		t.Devices[i].PipelineRank = i % nStages
	}
	return nil
}

// Enumerate detects available graphics adapters using DXGI or falls back to system primary adapter.
func Enumerate() (*Topology, error) {
	devices := make([]DeviceInfo, 0)

	// Attempt DXGI adapter enumeration via dxgi.dll
	dxgi := syscall.NewLazyDLL("dxgi.dll")
	createFactory1 := dxgi.NewProc("CreateDXGIFactory1")

	if err := createFactory1.Find(); err == nil {
		var factory uintptr
		// IID_IDXGIFactory1 = {770aae78-f26f-4dba-a829-253c83d1b387}
		iid := [16]byte{0x78, 0xae, 0x0a, 0x77, 0x6f, 0xf2, 0xba, 0x4d, 0xa8, 0x29, 0x25, 0x3c, 0x83, 0xd1, 0xb3, 0x87}
		r1, _, _ := createFactory1.Call(uintptr(unsafe.Pointer(&iid[0])), uintptr(unsafe.Pointer(&factory)))
		if r1 == 0 && factory != 0 {
			// IDXGIFactory1 vtable:
			// QueryInterface (0), AddRef (1), Release (2)
			// SetPrivateData (3), SetPrivateDataInterface (4), GetPrivateData (5), GetParent (6)
			// EnumAdapters (7), MakeWindowAssociation (8), GetWindowAssociation (9), CreateSwapChain (10), CreateSoftwareAdapter (11)
			// EnumAdapters1 (12)
			vtable := **(**[13]uintptr)(unsafe.Pointer(&factory))
			enumAdapters1 := vtable[12]

			for i := uint32(0); i < 8; i++ {
				var adapter uintptr
				res, _, _ := syscall.SyscallN(enumAdapters1, factory, uintptr(i), uintptr(unsafe.Pointer(&adapter)))
				if res != 0 || adapter == 0 {
					break
				}

				// DXGI_ADAPTER_DESC1 (vtable[8] is GetDesc1)
				adapterVtable := **(**[11]uintptr)(unsafe.Pointer(&adapter))
				getDesc1 := adapterVtable[10] // GetDesc1

				var desc struct {
					Description           [128]uint16
					VendorId              uint32
					DeviceId              uint32
					SubSysId              uint32
					Revision              uint32
					DedicatedVideoMemory  uintptr
					DedicatedSystemMemory uintptr
					SharedSystemMemory    uintptr
					AdapterLuid           uint64
					Flags                 uint32
				}

				rDesc, _, _ := syscall.SyscallN(getDesc1, adapter, uintptr(unsafe.Pointer(&desc)))
				if rDesc == 0 && (desc.Flags&2) == 0 { // Skip DXGI_ADAPTER_FLAG_SOFTWARE (2)
					name := syscall.UTF16ToString(desc.Description[:])
					dev := DeviceInfo{
						Index:        int(i),
						Name:         name,
						VendorID:     desc.VendorId,
						DeviceID:     desc.DeviceId,
						VRAMBytes:    uint64(desc.DedicatedVideoMemory),
						PipelineRank: -1,
					}
					// Derive stable 16-byte UUID from adapter vendor, device, and LUID
					h := sha256.New()
					fmt.Fprintf(h, "QADAV_GPU:%d:%d:%d:%s", desc.VendorId, desc.DeviceId, desc.AdapterLuid, name)
					sum := h.Sum(nil)
					copy(dev.UUID[:], sum[:16])

					devices = append(devices, dev)
				}

				// Release adapter (vtable[2])
				release := adapterVtable[2]
				_, _, _ = syscall.SyscallN(release, adapter)
			}

			// Release factory (vtable[2])
			factoryRelease := vtable[2]
			_, _, _ = syscall.SyscallN(factoryRelease, factory)
		}
	}

	// Fallback if no hardware devices enumerated via DXGI
	if len(devices) == 0 {
		h := sha256.Sum256([]byte("QADAV_PRIMARY_GPU_FALLBACK_DEVICE"))
		var uuid [16]byte
		copy(uuid[:], h[:16])
		devices = append(devices, DeviceInfo{
			Index:        0,
			Name:         "Default Compute Device (Single / Fallback)",
			VendorID:     0x8086,
			DeviceID:     0x1916,
			VRAMBytes:    2147483648,
			UUID:         uuid,
			PipelineRank: -1,
		})
	}

	top := &Topology{
		Devices:      devices,
		Interconnect: "shared_mem",
		SingleDevice: len(devices) <= 1,
	}

	return top, nil
}
