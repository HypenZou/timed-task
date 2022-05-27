package timedtask

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	tDB "timedtask/db"
	errors "timedtask/errors"

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
	db      *tDB.DB
	tw      *timingwheel.TimingWheel
	cache   *Cache
	codec   Serializable
	fn      func(msg interface{})
	onClean bool
	mu      sync.RWMutex
	path    string
}

func NewTimedTask(dbFilePath string, codec Serializable) (*TimedTask, error) {
	if err := makeFile(dbFilePath); err != nil {
		return nil, err
	}
	db, err := tDB.Open(dbFilePath)
	if err != nil {
		return nil, err
	}
	tw := timingwheel.NewTimingWheel(time.Millisecond, 20)
	tw.Start()
	cache := NewCache()
	timedtask := &TimedTask{db: db, codec: codec, tw: tw, path: dbFilePath, cache: cache}
	return timedtask, nil
}

func (timedtask *TimedTask) SetTask(fn func(msg interface{})) {
	timedtask.fn = fn
}

func (timedtask *TimedTask) AddTrigger(d time.Duration, stableMsg interface{}) error {
	if timedtask.fn == nil {
		return errors.ErrNoTask
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
	if timedtask.onClean {
		timedtask.cache.Put([]byte(fmt.Sprint(t)), msg, PUT)
	}
	// 已经超时则直接执行
	if time.Now().After(expire) {
		timedtask.doTask(stableMsg, []byte(fmt.Sprint(t)))
		return nil
	}
	timedtask.tw.AfterFunc(time.Until(expire), func() {
		timedtask.doTask(stableMsg, []byte(fmt.Sprint(t)))
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
		timedtask.tw.AfterFunc(time.Until(t), func() {
			timedtask.fn(msg)
			timedtask.db.Del([]byte(fmt.Sprint(t)))
		})
	}
	return nil
}

func (timedtask *TimedTask) Clean() error {
	newDb, err := tDB.Open("clean")
	if err != nil {
		return err
	}
	keys := timedtask.db.GetAll()
	timedtask.onClean = true
	for i := range keys {
		v, err := timedtask.db.Get([]byte(keys[i]))
		if err == errors.ErrNotFound {
			continue
		}
		newDb.Put([]byte(keys[i]), v)
	}
	for !timedtask.cache.IsEmpty() {
		key, value, method := timedtask.cache.Get()
		if method == PUT {
			newDb.Put(key, value)
		}
		if method == DEL {
			newDb.Del(key)
		}
	}
	time.Sleep(time.Second)
	timedtask.mu.Lock()
	timedtask.db = newDb
	os.RemoveAll(timedtask.path)
	os.Rename("clean", timedtask.path)
	timedtask.mu.Unlock()
	return nil
}

func (timedtask *TimedTask) doTask(stableMsg interface{}, key []byte) {
	timedtask.mu.RLock()
	defer timedtask.mu.RUnlock()
	timedtask.fn(stableMsg)
	timedtask.db.Del(key)
	if timedtask.onClean {
		timedtask.cache.Put(key, nil, DEL)
	}
}

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
