package test

import (
	"github.com/llyb120/vermouth"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestThreadLocal(t *testing.T) {
	tl := vermouth.NewThreadLocal()
	tl.Set("test")
	assert.Equal(t, "test", tl.Get())

	var wg sync.WaitGroup
	wg.Add(1)
	tl.Go(func() {
		defer wg.Done()
		assert.Equal(t, "test", tl.Get())

		tl.Set("test2")
		assert.Equal(t, "test2", tl.Get())
	})

	assert.Equal(t, "test", tl.Get())

	wg.Wait()
}
