package retry

import "time"

func Try(fn func() error) error {
	durations := [3]int{1, 3, 5}
	var err error
	err = fn()
	for try := 0; try < len(durations) && err != nil; try++ {
		time.Sleep(time.Duration(durations[try]) * time.Second)
		err = fn()
	}
	return err
}
