package command

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/logger"
	"github.com/hetianyi/gox/timer"
	"sync"
	"time"
)

var (
	testLock     = new(sync.Mutex)
	testFailed   = 0
	testSuccess  = 0
	resultBuffer *bufio.Writer
)

// handleTestUploadFile handles upload files by client cli.
func handleTestUploadFile() error {
	// initialize APIClient
	if err := initClient(); err != nil {
		logger.Fatal(err)
	}

	out, err := file.CreateFile("test-result.txt")
	if err != nil {
		logger.Fatal(err)
	}

	bufSize := 1 << 15
	resultBuffer = bufio.NewWriterSize(out, bufSize)

	showTestProgress(common.InitializedClientConfiguration.TestScale)

	startTime := gox.GetTimestamp(time.Now())
	waitGroup := sync.WaitGroup{}
	step := common.InitializedClientConfiguration.TestScale / common.InitializedClientConfiguration.TestThread

	for i := 1; i <= common.InitializedClientConfiguration.TestThread; i++ {
		if i == common.InitializedClientConfiguration.TestThread {
			go uploadTask((i-1)*step, common.InitializedClientConfiguration.TestScale, &waitGroup)
		} else {
			go uploadTask((i-1)*step, i*step, &waitGroup)
		}
	}

	waitGroup.Add(common.InitializedClientConfiguration.TestThread)
	waitGroup.Wait()

	endTime := gox.GetTimestamp(time.Now())
	// bug fixes: panic: runtime error: integer divide by zero
	if endTime-startTime < 1000 {
		endTime = startTime + 1000 // at least one sec.
	}
	resultBuffer.Flush()

	fmt.Println("+---------------------------+")
	fmt.Println("| total  :", common.InitializedClientConfiguration.TestScale)
	fmt.Println("| failed :", testFailed)
	fmt.Println("| time   :", (endTime-startTime)/1000, "s")
	fmt.Println("| average:", int64(common.InitializedClientConfiguration.TestScale)/((endTime-startTime)/1000), "/s")
	fmt.Println("+---------------------------+")
	return nil
}

func uploadTask(start int, end int, waitGroup *sync.WaitGroup) {
	for i := start; i < end; i++ {
		name := convert.IntToStr(i)
		data := []byte(name)
		size := int64(len(data))
		r := bytes.NewReader(data)
		//fmt.Println(gox.GetLongLongDateString(time.Now()), "  start upload")
		ret, err := client.Upload(r, size, group, common.InitializedClientConfiguration.PrivateUpload)
		//fmt.Println(gox.GetLongLongDateString(time.Now()), "  end   upload")
		if err != nil {
			logger.Error(err)
			updateTestCount(false)
		} else {
			bs, _ := json.MarshalIndent(ret, "", "  ")
			writeResult(bs)
			updateTestCount(true)
		}
	}
	waitGroup.Done()
}

func writeResult(bs []byte) {
	testLock.Lock()
	defer testLock.Unlock()

	resultBuffer.Write(bs)
}

func updateTestCount(s bool) {
	testLock.Lock()
	defer testLock.Unlock()

	switch s {
	case true:
		testSuccess++
		break
	case false:
		testFailed++
		break
	}
}

func showTestProgress(total int) {
	timer.Start(0, 0, time.Millisecond*20, func(t *timer.Timer) {
		fmt.Print(testSuccess, "/", total, "\r")
	})
}
