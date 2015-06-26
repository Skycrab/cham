package filter

import (
	// "fmt"
	"testing"
)

func TestAscFilter(t *testing.T) {
	r := New()
	r.Add("12")
	r.Add("13")
	r.Add("14")
	r.Add("15")
	r.Add("16")
	r.Add("34")
	r.Add("23")
	if r.Filter("1234121567") != "********67" {
		t.Error("filter ascii error")
	}
}

func TestRuneFilter(t *testing.T) {
	r := New()
	r.Add("毛片")
	r.Add("毛毛")
	if r.Filter("我是老毛") != "我是老毛" {
		t.Error("rune error")
	}
	if r.Filter("我是毛毛") != "我是**" {
		t.Error("rune 2 error")
	}

}
