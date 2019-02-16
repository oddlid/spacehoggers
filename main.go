// +build !make

package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	VERSION    string = "undef"
	COMMIT_ID  string = "undef"
	BUILD_DATE string = "undef"
)

type FInfo struct {
	Name string
	Path string
	Size uint64
}

// Implement sort interface
type BySize []FInfo

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

// HR for Human Readable sizes
func (f FInfo) HR() string {
	return bytes(f.Size)
}

func (f FInfo) RelPath() string {
	return filepath.Join(f.Path, f.Name)
}

func (f FInfo) String() string {
	return fmt.Sprintf("%10s  %s", f.HR(), f.RelPath())
}

// logn(), humanateBytes() and bytes() from:
// https://github.com/dustin/go-humanize/blob/master/bytes.go
func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

func humanateBytes(s uint64, base float64, sizes []string) string {
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

func bytes(s uint64) string {
	sizes := []string{" B", "kB", "MB", "GB", "TB", "PB", "EB"}
	return humanateBytes(s, 1000, sizes)
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			log.Debugf("dirSize(): %s", err)
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func listDir(rootDir string) ([]FInfo, error) {
	var fi []FInfo

	f, err := os.Open(rootDir)
	if err != nil {
		log.Debug("listDir(): Error opening rootDir")
		log.Error(err)
		return fi, err
	}
	defer f.Close()
	entries, err := f.Readdir(-1)
	if err != nil {
		log.Debug("listDir(): Error listing rootDir contents")
		log.Error(err)
		return fi, err
	}

	fi = make([]FInfo, 0, len(entries))

	for _, e := range entries {
		size, err := dirSize(filepath.Join(rootDir, e.Name()))
		if err != nil {
			log.Debug("listDir(): Got error back from dirSize()")
			log.Error(err)
			continue
		}
		fi = append(fi, FInfo{
			Name: e.Name(),
			Path: rootDir,
			Size: uint64(size),
		})
	}

	return fi, nil
}

func listFiles(rootDir string) ([]FInfo, error) {
	var fi []FInfo

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			//log.Debugf("Path: %q", path)
			fi = append(fi, FInfo{
				Name: info.Name(),
				Path: filepath.Dir(path), // strip filename from path
				Size: uint64(info.Size()),
			})
		}
		return err
	})

	return fi, err
}

func entryPoint(ctx *cli.Context) error {
	rootDir := ctx.String("root")
	rev := ctx.Bool("reverse")
	lim := ctx.Int("limit")
	all := ctx.Bool("all")

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

	if rev {
		sort.Sort(sort.Reverse(BySize(fi)))
	} else {
		sort.Sort(BySize(fi))
	}

	if len(fi) < lim || lim == 0 {
		lim = len(fi)
	}
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
			Name:  "log-level",
			Value: "info",
			Usage: "Log `level` (options: debug, info, warn, error, fatal, panic)",
		},
		cli.BoolFlag{
			Name:   "debug, d",
			Usage:  "Run in debug mode",
			EnvVar: "DEBUG",
		},
		cli.StringFlag{
			Name:  "root",
			Usage: "`DIR` to check",
			Value: ".", //os.Getenv("PWD"),
		},
		cli.IntFlag{
			Name:  "limit, l",
			Usage: "How many results to display",
			Value: 10,
		},
		cli.BoolFlag{
			Name:  "reverse, r",
			Usage: "Reverse order (smallest to largest)",
		},
		cli.BoolFlag{
			Name:  "all, a",
			Usage: "List all files instead of summarizing directories",
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
