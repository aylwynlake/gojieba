package jieba

import (
	"log"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestJiebaInstance(t *testing.T) {

	wg := &sync.WaitGroup{}
	n := 4
	insts := make([]*JiebaInstance, n)

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			insts[i] = NewJiebaInstance("")
		}(i)
	}
	wg.Wait()

	inst0 := insts[0]
	for i := 1; i < n; i++ {
		if !reflect.DeepEqual(inst0, insts[i]) {
			log.Fatalf("insts[%d] is not equal to inst0\n", i)
		}
	}

	for x := 0; x < 10; x++ {
		inst0.Reload()
		t, dur := inst0.LoadTime()
		log.Printf("%s %s\n", t.Format(time.RFC3339Nano), dur.String())
	}

}
