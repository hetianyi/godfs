package svc

import (
	"github.com/boltdb/bolt"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/logger"
	"github.com/hetianyi/gox/timer"
	"time"
)

// binlogPusher starts a timer job for pushing binlog to a tracker.
func binlogPusher(server *common.Server) {
	// allow 2 round failure synchronization
	timer.Start(time.Second*3, time.Second*10, 0, func(t *timer.Timer) {
		for true {
			// waiting for instanceId
			if server.InstanceId == "" {
				break
			}

			logger.Debug("reading binlog for tracker instance: ", server.InstanceId)

			fileIndex, offset, err := getPusherStatus(server.InstanceId)
			if err != nil {
				logger.Error("error reading pusher config: ", err)
				break
			}

			logger.Debug("tracker instanceId: ", server.InstanceId,
				", pusher status: binlog index is ", fileIndex, " and binlog offset is ", offset)

			bls, nOffset, err := writableBinlogManager.Read(fileIndex, offset, 10)
			if err != nil {
				logger.Error("error reading binlog: ", err)
				break
			}
			if bls != nil && len(bls) > 0 {
				if err := clientAPI.PushBinlog(server, bls); err != nil {
					logger.Error("error push binlog: ", err)
					break
				}
				logger.Debug(len(bls), " binlog pushed success")
			}
			if fileIndex == writableBinlogManager.GetCurrentIndex() && offset == nOffset {
				logger.Debug("no new binlog available")
				break
			}
			if writableBinlogManager.GetCurrentIndex() > fileIndex &&
				(bls == nil || len(bls) == 0) {
				fileIndex++
				nOffset = 0
				if err := setPusherStatus(server.InstanceId, fileIndex, nOffset); err != nil {
					logger.Error("error save binlog config for tracker instance: ", err)
				}
				break
			}
			if err := setPusherStatus(server.InstanceId, fileIndex, nOffset); err != nil {
				logger.Error("error save binlog config for tracker instance: ", err)
			}
		}
	})
}

// getPusherStatus gets current binlog push state of the tracker server.
func getPusherStatus(instanceId string) (fileIndex int, offset int64, err error) {
	configMap := common.GetConfigMap()
	pos, err := configMap.GetConfig("binlog_pos_" + instanceId)
	if err != nil {
		logger.Error("error load binlog position for tracker instance: ", instanceId)
		return
	}
	if pos == nil || len(pos) == 0 {
		configMap.PutConfig("binlog_pos_"+instanceId, []byte("0"))
		pos = []byte("0")
	}
	ind, err := configMap.GetConfig("binlog_index_" + instanceId)
	if err != nil {
		logger.Error("error load binlog file index for tracker instance: ", instanceId)
		return
	}
	if ind == nil || len(ind) == 0 {
		configMap.PutConfig("binlog_index_"+instanceId, []byte("0"))
		ind = []byte("0")
	}
	fileIndex, err = convert.StrToInt(string(ind))
	if err != nil {
		return
	}
	offset, err = convert.StrToInt64(string(pos))
	if err != nil {
		return
	}
	return
}

func setPusherStatus(instanceId string, fileIndex int, offset int64) error {
	configMap := common.GetConfigMap()
	return configMap.BatchUpdate(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(common.BUCKET_KEY_CONFIGMAP))
		err := b.Put([]byte("binlog_pos_"+instanceId), []byte(convert.Int64ToStr(offset)))
		if err != nil {
			logger.Error("error load binlog position for tracker instance: ", instanceId)
			return err
		}
		err = b.Put([]byte("binlog_index_"+instanceId), []byte(convert.IntToStr(fileIndex)))
		if err != nil {
			logger.Error("error load binlog file index for tracker instance: ", instanceId)
			return err
		}
		return nil
	})
}
