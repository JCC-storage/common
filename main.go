package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"gitlink.org.cn/cloudream/common/sdks/storage/cdsapi"
)

func main() {
	test1("http://121.36.5.116:32010")
	// test2("http://127.0.0.1:7890")
}

func test1(url string) {
	cli := cdsapi.NewClient(&cdsapi.Config{
		URL: url,
	})

	openLen, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		fmt.Println(err)
		return
	}

	readLen, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		fmt.Println(err)
		return
	}

	startTime := time.Now()
	obj, err := cli.Object().Download(cdsapi.ObjectDownload{
		UserID:   1,
		ObjectID: 470790,
		Offset:   0,
		Length:   &openLen,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Open time: %v\n", time.Since(startTime))

	startTime = time.Now()
	buf := make([]byte, readLen)
	_, err = io.ReadFull(obj.File, buf)
	fmt.Printf("Read time: %v\n", time.Since(startTime))
	if err != nil {
		fmt.Println(err)
		return
	}

	startTime = time.Now()
	obj.File.Close()
	fmt.Printf("Close time: %v\n", time.Since(startTime))
}

func test2(url string) {
	cli := cdsapi.NewClient(&cdsapi.Config{
		URL: url,
	})

	obj, err := cli.Object().Download(cdsapi.ObjectDownload{
		UserID:   1,
		ObjectID: 27151,
		Offset:   0,
		// Length:   &openLen,
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	f, err := os.Create("test.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	io.Copy(f, obj.File)
}
