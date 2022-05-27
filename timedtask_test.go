package timedtask

import (
	"fmt"
	"log"
	"testing"
	"time"
)

type serialize struct{}

func (s *serialize) Deserialize(data []byte) (interface{}, error) {
	return string(data), nil
}

func (s *serialize) Serialize(data interface{}) ([]byte, error) {
	if v, ok := data.(string); ok {
		return []byte(v), nil
	}
	return nil, SerializeErr
}

func TestTimedTask(t *testing.T) {
	timedtask, err := NewTimedTask("tmp", &serialize{})
	if err != nil {
		log.Printf("new timedtask faild : %s", err)
	}
	err = timedtask.Recover()
	if err != nil {
		log.Printf("recover faild : %s", err)
	}
	timedtask.SetTask(func(msg interface{}) {
		fmt.Println("打印:", msg)
	})
	//timedtask.AddTrigger(time.Second*34, "shit you")
	time.Sleep(time.Second * 100)

}
