package main

import (
	"fmt"
	"path/filepath"
	//"os"
)

/*
Tips for converting to MB:

sizeMB := float64(size) / 1024.0 / 1024.0
sizeMB = Round(sizeMB, 0.5, 2) // some homegrown func not defined in link

*/

// Taken from: https://stackoverflow.com/questions/32482673/how-to-get-directory-total-size
func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func main() {
	fmt.Println("vim-go")
}
