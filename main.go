//go:build !make
// +build !make

/*
2019-04-23:
There's still a lot of discrepancy between "du" on *nix vs this code. Should find what causes this.
*/

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type fileInfo struct {
	name      string
	path      string
	size      int64
	blocks    int64
	diskUsage int64
	//BlockSize int32
}
type byteSize float64
type bySize []fileInfo
type byDiskUsage []fileInfo

const (
	defaultBlockSize = 512
	defaultSort      = "size"
)

const (
	_           = iota
	KB byteSize = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

var (
	VERSION    = "undef"
	COMMIT_ID  = "undef"
	BUILD_DATE = "undef"
)

func newFileInfo(path string, ofi os.FileInfo) (fi fileInfo) {
	fi = fileInfo{
		name: ofi.Name(),
		path: path,
		size: ofi.Size(),
	}
	st := ofi.Sys()
	if st == nil {
		return
	}
	stt := st.(*syscall.Stat_t)
	fi.blocks = stt.Blocks
	//fi.BlockSize = stt.Blksize
	//llog := log.With().Int64("blocksize", stt.Blksize).Logger()
	//llog.Debug().Send()
	fi.diskUsage = fi.getDiskUsage()
	return
}

func (fi fileInfo) getDiskUsage() int64 {
	//return fi.Blocks * int64(fi.BlockSize) // There's something very wrong with this math...
	return fi.blocks * defaultBlockSize // is this really right for all (file)systems and disks?
}

// Implement sort interface
func (b bySize) Len() int {
	return len(b)
}

func (b bySize) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// Less has reversed logic, as default is to sort descending
func (b bySize) Less(i, j int) bool {
	return b[i].size > b[j].size
}

func (b byDiskUsage) Len() int {
	return len(b)
}

func (b byDiskUsage) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// Less has reversed logic, as default is to sort descending
func (b byDiskUsage) Less(i, j int) bool {
	return b[i].diskUsage > b[j].diskUsage
}

// HR for Human Readable sizes
func (f fileInfo) hrSize() string {
	//return bytes(f.Size)
	return hr(byteSize(f.size))
}

func (f fileInfo) hrDiskUsage() string {
	//return bytes(f.DiskUsage)
	return hr(byteSize(f.diskUsage))
}

func (f fileInfo) relPath() string {
	return filepath.Join(f.path, f.name)
}

func (f fileInfo) String() string {
	return fieldStr(f.hrSize(), f.hrDiskUsage(), f.relPath())
}

func fieldStr(size, used, path string) string {
	return fmt.Sprintf("%10s%12s  %s", size, used, path)
}

func hr(size byteSize) string {
	switch {
	case size > YB:
		return fmt.Sprintf("%.1f E", size/YB)
	case size > ZB:
		return fmt.Sprintf("%.1f E", size/ZB)
	case size > EB:
		return fmt.Sprintf("%.1f E", size/EB)
	case size > PB:
		return fmt.Sprintf("%.1f P", size/PB)
	case size > TB:
		return fmt.Sprintf("%.1f T", size/TB)
	case size > GB:
		return fmt.Sprintf("%.1f G", size/GB)
	case size > MB:
		return fmt.Sprintf("%.1f M", size/MB)
	case size > KB:
		return fmt.Sprintf("%.1f K", size/KB)
	default:
		return fmt.Sprintf("%.1f B", size)
	}
}

// logn(), humanateBytes() and bytes() from:
// https://github.com/dustin/go-humanize/blob/master/bytes.go
//func logn(n, b float64) float64 {
//	return math.Log(n) / math.Log(b)
//}
//
//func humanateBytes(s int64, base float64, sizes []string) string {
//	if s < 10 {
//		return fmt.Sprintf("%d B", s)
//	}
//	e := math.Floor(logn(float64(s), base))
//	suffix := sizes[int(e)]
//	val := math.Floor(float64(s)/math.Pow(base, e)*10+0.5) / 10
//	//	f := "%.0f %s"
//	//	if val < 10 {
//	//		f = "%.1f %s"
//	//	}
//	f := "%.1f %s"
//
//	return fmt.Sprintf(f, val, suffix)
//}
//
//func bytes(s int64) string {
//	//sizes := []string{" B", "KB", "MB", "GB", "TB", "PB", "EB"}
//	sizes := []string{"B", "K", "M", "G", "T", "P", "E"}
//	//return humanateBytes(s, 1000, sizes)
//	return humanateBytes(s, 1024, sizes)
//}

// getSizes() is just an attempt at a more efficient way of getting the size data
// for the dirSize func
func getSizes(info os.FileInfo) (size, diskUsage int64) {
	size = info.Size()
	st := info.Sys()
	if st == nil {
		return
	}
	stt := st.(*syscall.Stat_t)
	diskUsage = stt.Blocks * defaultBlockSize
	// So, what we probably need to do here, is to just add the remainder of stt.BlkSize,
	// not to multiply it by blocks, as that leads to a number waaaaay off the correct one.
	//diskUsage = stt.Blocks * stt.Blksize
	return
}

func dirSize(path string) (size, diskUsage int64, err error) {
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// With regards to actual disk usage, it would be better to include directories
		// here, but as this is for calculating bytes used by files, and the fact that we're
		// summarizing disk usage otherwise, it's probably better to do it this way here.
		//if !info.IsDir() {
		//	fi := NewFInfo(path, info)
		//	size += fi.Size
		//	diskUsage += fi.DiskUsage
		//}
		//fi := NewFInfo(path, info)
		//size += fi.Size
		//diskUsage += fi.DiskUsage
		s, d := getSizes(info)
		size += s
		diskUsage += d
		return err
	})
	return
}

func listDir(rootDir string) (fis []fileInfo, err error) {
	f, err := os.Open(rootDir)
	if err != nil {
		return
	}
	defer f.Close()
	entries, err := f.Readdir(-1)
	if err != nil {
		return
	}

	fis = make([]fileInfo, 0, len(entries))

	for _, e := range entries {
		size, diskUsage, err := dirSize(filepath.Join(rootDir, e.Name()))
		if err != nil {
			log.Error().Err(err).Send()
			continue
		}
		//fi := NewFInfo(rootDir, e)
		fi := fileInfo{
			name:      e.Name(),
			path:      rootDir,
			size:      size,
			diskUsage: diskUsage, // this is why we don't use NewFInfo here
		}
		fis = append(fis, fi)
	}

	return
}

func listFiles(rootDir string) (fis []fileInfo, err error) {
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			//log.Debugf("Path: %q", path)
			fi := newFileInfo(filepath.Dir(path), info)
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

	sortBySize := srt == defaultSort

	var fi []fileInfo
	var err error

	if all {
		fi, err = listFiles(rootDir)
	} else {
		fi, err = listDir(rootDir)
	}
	if err != nil {
		//return cli.NewExitError(err.Error(), 1)
		log.Error().Err(err).Send()
	}

	if sortBySize {
		if rev {
			sort.Sort(sort.Reverse(bySize(fi)))
		} else {
			sort.Sort(bySize(fi))
		}
	} else {
		if rev {
			sort.Sort(sort.Reverse(byDiskUsage(fi)))
		} else {
			sort.Sort(byDiskUsage(fi))
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

	app.Authors = []*cli.Author{
		{
			Name:  "Odd E. Ebbesen",
			Email: "oddebb@gmail.com",
		},
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "root",
			Aliases: []string{"R"},
			Usage:   "`DIR` to check",
			Value:   ".", //os.Getenv("PWD"),
		},
		&cli.BoolFlag{
			Name:    "all",
			Aliases: []string{"a"},
			Usage:   "List all files instead of summarizing directories",
		},
		&cli.StringFlag{
			Name:    "sort",
			Aliases: []string{"s"},
			Usage:   "Sort by `OPTION`: size or usage",
			Value:   "size",
		},
		&cli.BoolFlag{
			Name:    "reverse",
			Aliases: []string{"r"},
			Usage:   "Reverse order (smallest to largest)",
		},
		&cli.IntFlag{
			Name:    "limit",
			Aliases: []string{"l"},
			Usage:   "How many results to display",
			Value:   10,
		},
		&cli.StringFlag{
			Name:  "log-level",
			Value: "info",
			Usage: "Log `level` (options: debug, info, warn, error, fatal, panic)",
		},
		&cli.BoolFlag{
			Name:    "debug",
			Aliases: []string{"d"},
			Usage:   "Run in debug mode",
			EnvVars: []string{"DEBUG"},
		},
	}

	app.Before = func(c *cli.Context) error {
		zerolog.TimeFieldFormat = "2006-01-02T15:04:05.999-07:00"
		if c.Bool("debug") {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			if c.IsSet("log-level") {
				level, err := zerolog.ParseLevel(c.String("log-level"))
				if err != nil {
					log.Error().Err(err).Send()
				} else {
					zerolog.SetGlobalLevel(level)
				}
			} else {
				zerolog.SetGlobalLevel(zerolog.InfoLevel)
			}
		}
		return nil
	}

	app.Action = entryPoint
	app.Run(os.Args)
}
