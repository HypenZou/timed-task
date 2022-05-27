package timedtask

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"timedtask/minidb"

	"github.com/RussellLuo/timingwheel"
)

// type Stable interface {
// 	Put(key []byte, value []byte) error
// 	Get(key []byte) ([]byte, error)
// }

type Serializable interface {
	Serialize(data interface{}) ([]byte, error)
	Deserialize(data []byte) (interface{}, error)
}

type TimedTask struct {
	db    *minidb.MiniDB
	tw    *timingwheel.TimingWheel
	codec Serializable
	fn    func(msg interface{})
	mu    sync.RWMutex
}

// var _ Stable = (*minidb.MiniDB)(nil)

func NewTimedTask(dbFilePath string, codec Serializable) (*TimedTask, error) {
	if err := makeFile(dbFilePath); err != nil {
		return nil, err
	}
	db, err := minidb.Open(dbFilePath)
	if err != nil {
		return nil, err
	}
	tw := timingwheel.NewTimingWheel(time.Millisecond, 20)
	tw.Start()
	timedtask := &TimedTask{db: db, codec: codec, tw: tw}
	return timedtask, nil
}

func (timedtask *TimedTask) SetTask(fn func(msg interface{})) {
	timedtask.fn = fn
}

func (timedtask *TimedTask) AddTrigger(d time.Duration, stableMsg interface{}) error {
	if timedtask.fn == nil {
		return NoTaskErr
	}
	expire := time.Now().Add(d)
	t := time.Now().Add(d).UnixNano()
	msg, err := timedtask.codec.Serialize(stableMsg)
	if err != nil {
		return err
	}
	timedtask.mu.RLock()
	defer timedtask.mu.RUnlock()
	// 持久化
	timedtask.db.Put([]byte(fmt.Sprint(t)), msg)

	// 已经超时则直接执行
	if time.Now().After(expire) {
		timedtask.fn(stableMsg)
		timedtask.db.Del([]byte(fmt.Sprint(t)))
		return nil
	}
	timedtask.tw.AfterFunc(time.Until(expire), func() {
		timedtask.fn(stableMsg)
		timedtask.db.Del([]byte(fmt.Sprint(t)))
	})
	return nil
}

func (timedtask *TimedTask) Recover() error {
	keys := timedtask.db.GetAll()
	for i := range keys {
		v, err := timedtask.db.Get([]byte(keys[i]))
		if err != nil {
			log.Printf("recover error: %s", err)
			continue
		}
		digitt, _ := strconv.ParseInt(keys[i], 10, 64)
		msg, err := timedtask.codec.Deserialize(v)
		if err != nil {
			return err
		}
		t := time.Unix(digitt/1e9, digitt-int64(int(digitt/1e9))*1e9)
		if time.Now().After(t) {
			continue
		}
		fmt.Println(t)
		timedtask.tw.AfterFunc(time.Until(t), func() {
			timedtask.fn(msg)
			timedtask.db.Del([]byte(fmt.Sprint(t)))
		})
	}
	return nil
}

// func (timedtask *TimedTask) Clean() error {
// 	timedtask.mu.Lock()
// 	defer timedtask.mu.Unlock()
// 	keys := timedtask.db.GetAll()
// 	if err != nil {
// 		return err
// 	}
// 	timedtask.onClean = true
// 	for i := range keys {
// 		timedtask.db.Get([]byte(keys[i]))
// 		newDb.Put([]byte(keys[i]))
// 	}
// }

func makeFile(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		// 创建文件夹
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Printf("mkdir failed![%v]\n", err)
		} else {
			return nil
		}
	}
	return err
}
