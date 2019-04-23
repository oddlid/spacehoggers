// +build !make

package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	//"golang.org/x/sys/unix"
)

const (
	DEF_BLK_SIZE = 512
	DEF_SORT     = "size"
)

var (
	VERSION    string = "undef"
	COMMIT_ID  string = "undef"
	BUILD_DATE string = "undef"
)

type FInfo struct {
	Name      string
	Path      string
	Size      int64
	Blocks    int64
	BlockSize int32
	DiskUsage int64
}

func NewFInfo(name, path string, ofi os.FileInfo) (fi FInfo) {
	fi = FInfo{
		Name: name,
		Path: path,
		Size: ofi.Size(),
	}
	st := ofi.Sys()
	if st == nil {
		return
	}
	stt := st.(*syscall.Stat_t)
	fi.Blocks = stt.Blocks
	fi.BlockSize = stt.Blksize
	fi.DiskUsage = fi.diskUsage()
	return
}

func (fi FInfo) diskUsage() int64 {
	//return fi.Blocks * int64(fi.BlockSize) // There's something very wrong with this math...
	return fi.Blocks * DEF_BLK_SIZE // is this really right for all (file)systems and disks?
}

// Implement sort interface
type BySize []FInfo
type ByDiskUsage []FInfo

func (b BySize) Len() int {
	return len(b)
}

func (b BySize) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// Less has reversed logic, as default is to sort descending
func (b BySize) Less(i, j int) bool {
	return b[i].Size > b[j].Size
}

func (b ByDiskUsage) Len() int {
	return len(b)
}

func (b ByDiskUsage) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// Less has reversed logic, as default is to sort descending
func (b ByDiskUsage) Less(i, j int) bool {
	return b[i].DiskUsage > b[j].DiskUsage
}

// HR for Human Readable sizes
func (f FInfo) HRSize() string {
	return bytes(f.Size)
}

func (f FInfo) HRDiskUsage() string {
	return bytes(f.DiskUsage)
}

func (f FInfo) RelPath() string {
	return filepath.Join(f.Path, f.Name)
}

func (f FInfo) String() string {
	return fieldStr(f.HRSize(), f.HRDiskUsage(), f.RelPath())
}

func fieldStr(size, used, path string) string {
	return fmt.Sprintf("%10s%12s  %s", size, used, path)
}

// logn(), humanateBytes() and bytes() from:
// https://github.com/dustin/go-humanize/blob/master/bytes.go
func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

func humanateBytes(s int64, base float64, sizes []string) string {
	if s < 10 {
		return fmt.Sprintf("%d B", s)
	}
	e := math.Floor(logn(float64(s), base))
	suffix := sizes[int(e)]
	val := math.Floor(float64(s)/math.Pow(base, e)*10+0.5) / 10
	//	f := "%.0f %s"
	//	if val < 10 {
	//		f = "%.1f %s"
	//	}
	f := "%.1f %s"

	return fmt.Sprintf(f, val, suffix)
}

func bytes(s int64) string {
	//sizes := []string{" B", "KB", "MB", "GB", "TB", "PB", "EB"}
	sizes := []string{"B", "K", "M", "G", "T", "P", "E"}
	//return humanateBytes(s, 1000, sizes)
	return humanateBytes(s, 1024, sizes)
}

func dirSize(path string) (size, diskUsage int64, err error) {
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Debugf("dirSize(): %s", err)
			return err
		}
		// With regards to actual disk usage, it would be better to include directories
		// here, but as this is for calculating bytes used by files, and the fact that we're
		// summarizing disk usage otherwise, it's probably better to do it this way here.
		if !info.IsDir() {
			fi := NewFInfo(info.Name(), path, info)
			size += fi.Size
			diskUsage += fi.DiskUsage
		}
		return err
	})
	return
}

func listDir(rootDir string) (fis []FInfo, err error) {
	f, err := os.Open(rootDir)
	if err != nil {
		log.Debug("listDir(): Error opening rootDir")
		log.Error(err)
		return
	}
	defer f.Close()
	entries, err := f.Readdir(-1)
	if err != nil {
		log.Debug("listDir(): Error listing rootDir contents")
		log.Error(err)
		return
	}

	fis = make([]FInfo, 0, len(entries))

	for _, e := range entries {
		size, diskUsage, err := dirSize(filepath.Join(rootDir, e.Name()))
		if err != nil {
			log.Debug("listDir(): Got error back from dirSize()")
			log.Error(err)
			continue
		}
		//fi := NewFInfo(e.Name(), rootDir, e)
		fi := FInfo{
			Name:      e.Name(),
			Path:      rootDir,
			Size:      size,
			DiskUsage: diskUsage, // this is why we don't use NewFInfo here
		}
		fis = append(fis, fi)
	}

	return
}

func listFiles(rootDir string) (fis []FInfo, err error) {
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			//log.Debugf("Path: %q", path)
			fi := NewFInfo(info.Name(), filepath.Dir(path), info)
			fis = append(fis, fi)
		}
		return err
	})

	return
}

func entryPoint(ctx *cli.Context) error {
	rootDir := ctx.String("root")
	srt := ctx.String("sort")
	rev := ctx.Bool("reverse")
	all := ctx.Bool("all")
	lim := ctx.Int("limit")

	bySize := srt == DEF_SORT

	var fi []FInfo
	var err error

	if all {
		fi, err = listFiles(rootDir)
	} else {
		fi, err = listDir(rootDir)
	}
	if err != nil {
		//return cli.NewExitError(err.Error(), 1)
		log.Error(err)
	}

	if bySize {
		if rev {
			sort.Sort(sort.Reverse(BySize(fi)))
		} else {
			sort.Sort(BySize(fi))
		}
	} else {
		if rev {
			sort.Sort(sort.Reverse(ByDiskUsage(fi)))
		} else {
			sort.Sort(ByDiskUsage(fi))
		}
	}

	if len(fi) < lim || lim == 0 {
		lim = len(fi)
	}

	fmt.Println(fieldStr("Size", "Usage", "Path"))
	fmt.Println("----------------------------")

	for i := 0; i < lim; i++ {
		fmt.Println(fi[i].String())
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "spacehoggers"
	app.Usage = "Find biggest/smallest files/dirs"
	app.Copyright = "(c) 2019 Odd Eivind Ebbesen"
	app.Version = fmt.Sprintf("%s_%s (Compiled: %s)", VERSION, COMMIT_ID, BUILD_DATE)
	app.Compiled, _ = time.Parse(time.RFC3339, BUILD_DATE)

	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Odd E. Ebbesen",
			Email: "oddebb@gmail.com",
		},
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "root, R",
			Usage: "`DIR` to check",
			Value: ".", //os.Getenv("PWD"),
		},
		cli.BoolFlag{
			Name:  "all, a",
			Usage: "List all files instead of summarizing directories",
		},
		cli.StringFlag{
			Name:  "sort, s",
			Usage: "Sort by `OPTION`: size or usage",
			Value: "size",
		},
		cli.BoolFlag{
			Name:  "reverse, r",
			Usage: "Reverse order (smallest to largest)",
		},
		cli.IntFlag{
			Name:  "limit, l",
			Usage: "How many results to display",
			Value: 10,
		},
		cli.StringFlag{
			Name:  "log-level",
			Value: "info",
			Usage: "Log `level` (options: debug, info, warn, error, fatal, panic)",
		},
		cli.BoolFlag{
			Name:   "debug, d",
			Usage:  "Run in debug mode",
			EnvVar: "DEBUG",
		},
	}

	app.Before = func(c *cli.Context) error {
		log.SetOutput(os.Stderr)
		level, err := log.ParseLevel(c.String("log-level"))
		if err != nil {
			log.Fatal(err.Error())
		}
		log.SetLevel(level)
		if !c.IsSet("log-level") && !c.IsSet("l") && c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}
		log.SetFormatter(&log.TextFormatter{
			DisableTimestamp: false,
			FullTimestamp:    true,
		})
		return nil
	}

	app.Action = entryPoint
	app.Run(os.Args)
}
