package main

import (
	"fmt"
	"os"
	"testing"
)

func CreateMockImport(t *testing.T) {
	f, err := os.Open("import.txt")
	if err != nil {
		fmt.Println("", err)
		return
	}
	defer f.Close()
	for i := 0; i < 1000; i++ {
		text := fmt.Sprintf("%d_asdpfhisdfasdfasdfwe3efsfsdafasdfsdfasdfasdfasdfasdfasdfasdfasdxddasdfasd /Volumes/new-apfs/devnet/a.car t01000\n", i)
		f.Write([]byte(text))
	}
}
