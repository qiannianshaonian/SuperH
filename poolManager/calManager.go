package poolManager

import (
	"fmt"
	"sync"
)

var (
	calPoolchan chan string
	calPoolMap  sync.Map
)

func CalPoolInit() {
	calPoolchan = make(chan string, 500)
	for i := int64(0); i < 500; i++ {
		calPoolId := fmt.Sprintf("calPool_%d", i)
		go newCal(calPoolId)
	}
}
