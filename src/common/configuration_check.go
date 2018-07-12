package common

import (
	"strconv"
	"util/logger"
	"strings"
	"path/filepath"
	"os"
	"util/file"
	"bytes"
)

// check configuration file parameter.
// if check failed, system will shutdown.
// runWith:
//		1: storage server
//		2: tracker server
func check(m map[string] string, runWith int) {
	// check: bind_address
	//bind_address := m["bind_address"]

	// check port
	port, e := strconv.Atoi(m["port"])
	if e == nil {
		if port <= 0 || port > 65535 {
			logger.Fatal("invalid port range:", m["port"])
		}
	} else {
		logger.Fatal("invalid port ", m["port"], ":", e)
	}

	// check base_path
	base_path := strings.TrimSpace(m["base_path"])
	if base_path == "" {
		abs,_ := filepath.Abs(os.Args[0])
		parent, _ := filepath.Split(abs)
		finalPath := file.FixPath(parent + string(filepath.Separator) + "godfs")
		logger.Info("base_path not set, use", finalPath)
		if file.Exists(finalPath) && file.IsFile(finalPath) {
			logger.Fatal("could not create base path:", finalPath)
		}

		if !file.Exists(finalPath) {
			e := file.CreateDir(finalPath)
			if e != nil {
				logger.Fatal("could not create base path:", finalPath)
			}
		}
		m["base_path"] = finalPath
	}


	// check secret
	m["secret"] = strings.TrimSpace(m["secret"])

	if runWith == 1 {

		// check trackers
		trackers := strings.TrimSpace(m["trackers"])
		_ts := strings.Split(trackers, ",")
		var bytebuff bytes.Buffer
		for i := range _ts {
			strS := strings.TrimSpace(_ts[i])
			if strS == "" {
				continue
			}
			bytebuff.WriteString(strS)
			if i < len(_ts) {
				bytebuff.WriteString(",")
			}
		}
		m["trackers"] = string(bytebuff.Bytes())
		//--
	}

	if runWith == 2 {

	}

}

