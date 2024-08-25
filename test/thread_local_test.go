package test

import (
	"testing"
	"github.com/llyb120/vermouth"
	"github.com/stretchr/testify/assert"
	"sync"
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
	})

	wg.Wait()
}