package handler

import (
	"fmt"
	"gen-piece-commitment/inited"
	"testing"
	"time"
)

func TestDBHandler_GetDeals(t *testing.T) {

}

func TestRedisLock(t *testing.T) {
	inited.InitApp("/Volumes/new-apfs/projects/gen-piece-commitment/config.toml")
	work := func(w string, n int) {
		key := fmt.Sprintf("asdfd-%v", n)
		ok, err := inited.Redis.Lock(key, 00, 1*time.Second)
		if err != nil {
			fmt.Println("lock err: ", err.Error())
			return
		}
		if ok {
			time.Sleep(200 * time.Millisecond)

			fmt.Println("doing ", w, n)
		}
		inited.Redis.Unlock(key)
	}
	for i := 0; i < 10; i++ {
		n := i
		go work("a", n)
		go work("b", n)
	}

	time.Sleep(30 * time.Second)
}
