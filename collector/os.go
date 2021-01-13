// +build windows

package collector

import (
	"fmt"
	"time"

	"github.com/prometheus-community/windows_exporter/headers/custom"
	"github.com/prometheus-community/windows_exporter/headers/netapi32"
	"github.com/prometheus-community/windows_exporter/headers/psapi"
	"github.com/prometheus-community/windows_exporter/headers/sysinfoapi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	registerCollector("os", NewOSCollector)
}

// A OSCollector is a Prometheus collector for WMI metrics
type OSCollector struct {
	OSInformation           *prometheus.Desc
	PhysicalMemoryFreeBytes *prometheus.Desc
	PagingFreeBytes         *prometheus.Desc
	VirtualMemoryFreeBytes  *prometheus.Desc
	ProcessesLimit          *prometheus.Desc
	ProcessMemoryLimitBytes *prometheus.Desc
	Processes               *prometheus.Desc
	Users                   *prometheus.Desc
	PagingLimitBytes        *prometheus.Desc
	VirtualMemoryBytes      *prometheus.Desc
	VisibleMemoryBytes      *prometheus.Desc
	Time                    *prometheus.Desc
	Timezone                *prometheus.Desc
}

// NewOSCollector ...
func NewOSCollector() (Collector, error) {
	const subsystem = "os"

	return &OSCollector{
		OSInformation: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "info"),
			"OperatingSystem.Caption, OperatingSystem.Version",
			[]string{"product", "version"},
			nil,
		),
		PagingLimitBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "paging_limit_bytes"),
			"OperatingSystem.SizeStoredInPagingFiles",
			nil,
			nil,
		),
		PagingFreeBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "paging_free_bytes"),
			"OperatingSystem.FreeSpaceInPagingFiles",
			nil,
			nil,
		),
		PhysicalMemoryFreeBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "physical_memory_free_bytes"),
			"OperatingSystem.FreePhysicalMemory",
			nil,
			nil,
		),
		Time: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "time"),
			"OperatingSystem.LocalDateTime",
			nil,
			nil,
		),
		Timezone: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "timezone"),
			"OperatingSystem.LocalDateTime",
			[]string{"timezone"},
			nil,
		),
		Processes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "processes"),
			"OperatingSystem.NumberOfProcesses",
			nil,
			nil,
		),
		ProcessesLimit: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "processes_limit"),
			"OperatingSystem.MaxNumberOfProcesses",
			nil,
			nil,
		),
		ProcessMemoryLimitBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "process_memory_limix_bytes"),
			"OperatingSystem.MaxProcessMemorySize",
			nil,
			nil,
		),
		Users: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "users"),
			"OperatingSystem.NumberOfUsers",
			nil,
			nil,
		),
		VirtualMemoryBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "virtual_memory_bytes"),
			"OperatingSystem.TotalVirtualMemorySize",
			nil,
			nil,
		),
		VisibleMemoryBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "visible_memory_bytes"),
			"OperatingSystem.TotalVisibleMemorySize",
			nil,
			nil,
		),
		VirtualMemoryFreeBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "virtual_memory_free_bytes"),
			"OperatingSystem.FreeVirtualMemory",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *OSCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting os metrics:", desc, err)
		return err
	}
	return nil
}

// Win32_OperatingSystem docs:
// - https://msdn.microsoft.com/en-us/library/aa394239 - Win32_OperatingSystem class
type Win32_OperatingSystem struct {
	Caption                 string
	FreePhysicalMemory      uint64
	FreeSpaceInPagingFiles  uint64
	FreeVirtualMemory       uint64
	LocalDateTime           time.Time
	MaxNumberOfProcesses    uint32
	MaxProcessMemorySize    uint64
	NumberOfProcesses       uint32
	NumberOfUsers           uint32
	SizeStoredInPagingFiles uint64
	TotalVirtualMemorySize  uint64
	TotalVisibleMemorySize  uint64
	Version                 string
}

func (c *OSCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	/*var dst []Win32_OperatingSystem
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}*/

	product, buildNum := custom.GetProductDetails()

	nwgi, _, err := netapi32.NetWkstaGetInfo()
	if err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.OSInformation,
		prometheus.GaugeValue,
		1.0,
		fmt.Sprintf("Microsoft %s", product), // Caption
		fmt.Sprintf("%d.%d.%s", nwgi.Wki102_ver_major, nwgi.Wki102_ver_minor, buildNum), // Version
	)

	gmse, err := sysinfoapi.GlobalMemoryStatusEx()
	if err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.PhysicalMemoryFreeBytes,
		prometheus.GaugeValue,
		float64(gmse.UllAvailPhys),
	)

	currentTime := time.Now()

	ch <- prometheus.MustNewConstMetric(
		c.Time,
		prometheus.GaugeValue,
		float64(currentTime.Unix()),
	)

	timezoneName, _ := currentTime.Zone()

	ch <- prometheus.MustNewConstMetric(
		c.Timezone,
		prometheus.GaugeValue,
		1.0,
		timezoneName,
	)

	ch <- prometheus.MustNewConstMetric(
		c.PagingFreeBytes,
		prometheus.GaugeValue,
		float64(1234567),
		// Cannot find a way to get this without WMI.
		// Can get from CIM_OperatingSystem which is where WMI gets it from, but I can't figure out how to access this from cimwin32.dll
		// https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/cim-operatingsystem#properties
	)

	ch <- prometheus.MustNewConstMetric(
		c.VirtualMemoryFreeBytes,
		prometheus.GaugeValue,
		float64(gmse.UllAvailPageFile),
	)

	// Windows has no defined limit, and is based off available resources. This currently isn't calculated by WMI and is set to default value.
	// https://techcommunity.microsoft.com/t5/windows-blog-archive/pushing-the-limits-of-windows-processes-and-threads/ba-p/723824
	// https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-operatingsystem
	ch <- prometheus.MustNewConstMetric(
		c.ProcessesLimit,
		prometheus.GaugeValue,
		float64(4294967295),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ProcessMemoryLimitBytes,
		prometheus.GaugeValue,
		float64(gmse.UllTotalVirtual),
	)

	gpi, err := psapi.GetLPPerformanceInfo()
	if err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.Processes,
		prometheus.GaugeValue,
		float64(gpi.ProcessCount),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Users,
		prometheus.GaugeValue,
		float64(nwgi.Wki102_logged_on_users),
	)

	fsipf, err := custom.GetSizeStoredInPagingFiles()
	if err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.PagingLimitBytes,
		prometheus.GaugeValue,
		float64(fsipf),
	)

	ch <- prometheus.MustNewConstMetric(
		c.VirtualMemoryBytes,
		prometheus.GaugeValue,
		float64(gmse.UllTotalPageFile),
	)

	ch <- prometheus.MustNewConstMetric(
		c.VisibleMemoryBytes,
		prometheus.GaugeValue,
		float64(gmse.UllTotalPhys),
	)

	return nil, nil
}
