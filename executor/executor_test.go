package executor

import (
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
)

func TestThreadExecutor(t *testing.T) {
	var cnt int32
	executor := NewThreadExecutor(4)

	tasks := make([]TaskResult, 0, 100)
	for i := 0; i != 100; i++ {
		task := executor.Execute(func() {
			atomic.AddInt32(&cnt, 1)
		})
		tasks = append(tasks, task)
	}
	executor.Start()
	for _, task := range tasks {
		task.Wait()
	}
	assert.EqualValues(t, 100, cnt)
	executor.Stop()
}
