package db

import (
	"fmt"
	"testing"
)

func TestRecover(t *testing.T) {
	db, _ := Open("tmp")
	//db.Put([]byte("test"), []byte("value"))
	res, err := db.Get([]byte("test"))
	fmt.Println(string(res), err)
}
