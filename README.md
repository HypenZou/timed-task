# timed-task

It's a library helps you to add timed tasks in persistent way

here is the simplest demo:

```go 


type serialize struct{}

func (s *serialize) Deserialize(data []byte) (interface{}, error) {
	digitt, _ := strconv.ParseInt(string(data), 10, 64)
	t := time.Unix(digitt/1e9, digitt-int64(int(digitt/1e9))*1e9)
	return t, nil
}

func (s *serialize) Serialize(data interface{}) ([]byte, error) {
	if v, ok := data.(time.Time); ok {
		t := v.UnixNano()
		return []byte(strconv.FormatInt(t, 10)), nil
	}
	return nil, errors.ErrSerialize
}

func TestTimedTask(t *testing.T) {
	var se serialize
	timedtask, err := NewTimedTask("tmp", &se)
	if err != nil {
		log.Printf("new timedtask faild : %s", err)
	}
	err = timedtask.Recover()
	if err != nil {
		log.Printf("recover faild : %s", err)
	}
	timedtask.SetTask(func(msg interface{}) {
		t := time.Now()
		fmt.Println("触发误差:", t.Sub(msg.(time.Time)))
	})
	for i := 0; i < 10000; i++ {
		t := time.Millisecond * time.Duration(10*rand.Intn(100)+1)
		timedtask.AddTrigger(t, time.Now().Add(t))
		time.Sleep(time.Millisecond * 10)
	}

}

```