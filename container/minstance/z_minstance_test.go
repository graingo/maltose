package minstance_test

import (
	"strconv"
	"sync"
	"testing"

	"github.com/graingo/maltose/container/minstance"
	"github.com/stretchr/testify/assert"
)

func TestContainer_New(t *testing.T) {
	c := minstance.New()
	assert.NotNil(t, c)
	assert.Equal(t, 0, c.Count())
}

func TestContainer_SetAndGet(t *testing.T) {
	c := minstance.New()
	c.Set("name", "John")
	assert.Equal(t, "John", c.Get("name"))
	assert.Nil(t, c.Get("age"))
}

func TestContainer_GetOrSetFunc(t *testing.T) {
	c := minstance.New()
	val := c.GetOrSetFunc("name", func() any {
		return "John"
	})
	assert.Equal(t, "John", val)

	val = c.GetOrSetFunc("name", func() any {
		return "Doe"
	})
	assert.Equal(t, "John", val)
}

func TestContainer_Remove(t *testing.T) {
	c := minstance.New()
	c.Set("name", "John")
	assert.Equal(t, "John", c.Get("name"))

	c.Remove("name")
	assert.Nil(t, c.Get("name"))
}

func TestContainer_AllAndCount(t *testing.T) {
	c := minstance.New()
	c.Set("name", "John")
	c.Set("age", 30)

	assert.Equal(t, 2, c.Count())
	all := c.All()
	assert.Len(t, all, 2)
	// Note: The order of elements in All() is not guaranteed
	assert.Contains(t, all, "John")
	assert.Contains(t, all, 30)
}

func TestContainer_Concurrency(t *testing.T) {
	c := minstance.New()
	var wg sync.WaitGroup
	count := 100

	// Concurrent Set
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()
			c.Set("key"+strconv.Itoa(i), i)
		}(i)
	}
	wg.Wait()
	assert.Equal(t, count, c.Count())

	// Concurrent Get
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()
			val := c.Get("key" + strconv.Itoa(i))
			assert.Equal(t, i, val)
		}(i)
	}
	wg.Wait()

	// Concurrent GetOrSetFunc
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()
			c.GetOrSetFunc("getorset"+strconv.Itoa(i), func() any {
				return i
			})
		}(i)
	}
	wg.Wait()

	for i := 0; i < count; i++ {
		val := c.Get("getorset" + strconv.Itoa(i))
		assert.Equal(t, i, val)
	}

}
