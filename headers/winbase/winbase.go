package winbase

import "golang.org/x/sys/windows"

var (
	kernel32                    = windows.NewLazySystemDLL("kernel32.dll")
	procGetActiveProcessorCount = kernel32.NewProc("GetActiveProcessorCount")
)

const ALL_PROCESSOR_GROUPS uint16 = 0xffff

func GetActiveProcessorCount(groupNumber uint16) (int, error) {
	r1, _, err := procGetActiveProcessorCount.Call()
	if r1 == 0 {
		return 0, err
	}
	return int(r1), nil
}
