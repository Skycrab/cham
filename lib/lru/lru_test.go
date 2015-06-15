package lru

import (
	"fmt"
	"testing"
)

func TestLru(t *testing.T) {
	c := New(2, nil)
	c.Add(1, "hello")
	c.Add(2, "world")
	if d, ok := c.Get(1); ok {
		if d.(string) != "hello" {
			t.Error("error")
		}
	} else {
		t.Error("error")
	}
	c.Add(3, "overload")
	if e, ok := c.Get(2); ok || e != nil {
		t.Error("remove oldest error")
	}
	if c.Len() != 2 {
		t.Error("len error")
	}
	fmt.Println("end")
}
