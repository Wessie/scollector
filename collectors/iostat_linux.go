package collectors

import (
	"strconv"
	"strings"

	"github.com/StackExchange/scollector/opentsdb"
	"github.com/StackExchange/slog"
)

func init() {
	collectors = append(collectors, &IntervalCollector{F: c_iostat_linux})
}

var FIELDS_DISK = []string{
	"read_requests",       // Total number of reads completed successfully.
	"read_merged",         // Adjacent read requests merged in a single req.
	"read_sectors",        // Total number of sectors read successfully.
	"msec_read",           // Total number of ms spent by all reads.
	"write_requests",      // Total number of writes completed successfully.
	"write_merged",        // Adjacent write requests merged in a single req.
	"write_sectors",       // Total number of sectors written successfully.
	"msec_write",          // Total number of ms spent by all writes.
	"ios_in_progress",     // Number of actual I/O requests currently in flight.
	"msec_total",          // Amount of time during which ios_in_progress >= 1.
	"msec_weighted_total", // Measure of recent I/O completion time and backlog.
}

var FIELDS_PART = []string{
	"read_issued",
	"read_sectors",
	"write_issued",
	"write_sectors",
}

func c_iostat_linux() opentsdb.MultiDataPoint {
	var md opentsdb.MultiDataPoint
	readLine("/proc/diskstats", func(s string) {
		values := strings.Fields(s)
		if len(values) < 4 {
			return
		} else if values[3] == "0" {
			// Skip disks that haven't done a single read.
			return
		}
		metric := "linux.disk.part."
		i0, _ := strconv.Atoi(values[0])
		i1, _ := strconv.Atoi(values[1])
		if i1%16 == 0 && i0 > 1 {
			metric = "linux.disk."
		}
		device := values[2]
		if len(values) == 14 {
			var read_sectors, msec_read, write_sectors, msec_write float64
			for i, v := range values[3:] {
				switch FIELDS_DISK[i] {
				case "read_sectors":
					read_sectors, _ = strconv.ParseFloat(v, 64)
				case "msec_read":
					msec_read, _ = strconv.ParseFloat(v, 64)
				case "write_sectors":
					write_sectors, _ = strconv.ParseFloat(v, 64)
				case "msec_write":
					msec_write, _ = strconv.ParseFloat(v, 64)
				}
				Add(&md, metric+FIELDS_DISK[i], v, opentsdb.TagSet{"dev": device})
			}
			if read_sectors != 0 && msec_read != 0 {
				Add(&md, metric+"time_per_read", read_sectors/msec_read, opentsdb.TagSet{"dev": device})
			}
			if write_sectors != 0 && msec_write != 0 {
				Add(&md, metric+"time_per_write", write_sectors/msec_write, opentsdb.TagSet{"dev": device})
			}
		} else if len(values) == 7 {
			for i, v := range values[3:] {
				Add(&md, metric+FIELDS_PART[i], v, opentsdb.TagSet{"dev": device})
			}
		} else {
			slog.Infoln("iostat: cannot parse")
		}
	})
	return md
}
