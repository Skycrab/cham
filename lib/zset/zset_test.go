package zset

import (
	"fmt"
	"testing"
)

func equal(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func assert(t *testing.T, ok bool, s string) {
	if !ok {
		t.Error(s)
	}
}

func TestBase(t *testing.T) {
	z := New()
	assert(t, z.Count() == 0, "empty Count error")
	z.Add(1, "12")
	z.Add(1, "32")
	assert(t, z.Count() == 2, "not empty Count error")
	var score float64
	var ex bool
	score, ex = z.Score("12")
	assert(t, score == 1, "Score error")
	z.Add(2, "12")
	assert(t, z.Count() == 2, "after add duplicate Count error")
	score, ex = z.Score("12")
	assert(t, score == 2, "after add Score error")
	z.Rem("12")
	assert(t, z.Count() == 1, "after rem Count error")
	score, ex = z.Score("12")
	assert(t, ex == false, "not exist Score error")
	fmt.Println("")
}

func TestRangeByScore(t *testing.T) {
	z := New()
	z.Add(2, "22")
	z.Add(1, "11")
	z.Add(3, "33")
	s := "TestRangeByScore error"
	assert(t, equal(z.RangeByScore(2, 3), []string{"22", "33"}), s)
	assert(t, equal(z.RangeByScore(0, 5), []string{"11", "22", "33"}), s)
	assert(t, equal(z.RangeByScore(10, 5), []string{}), s)
	assert(t, equal(z.RangeByScore(10, 0), []string{"33", "22", "11"}), s)
}

func TestRange(t *testing.T) {
	z := New()
	z.Add(100.1, "1")
	z.Add(100.9, "9")
	z.Add(100.5, "5")
	assert(t, equal(z.Range(1, 3), []string{"1", "5", "9"}), "Range1 error")
	assert(t, equal(z.Range(3, 1), []string{"9", "5", "1"}), "Range2 error")
	assert(t, equal(z.RevRange(1, 2), []string{"9", "5"}), "RevRange1 error")
	assert(t, equal(z.RevRange(3, 2), []string{"1", "5"}), "RevRange2 error")

}

func TestRank(t *testing.T) {
	z := New()
	assert(t, z.Rank("kehan") == 0, "Rank empty error")
	z.Add(1111.1111, "kehan")
	assert(t, z.Rank("kehan") == 1, "Rank error")
	z.Add(222.2222, "lwy")
	assert(t, z.Rank("kehan") == 2, "Rank 2 error")
	assert(t, z.RevRank("kehan") == 1, "RevRank error")
}

func TestLimit(t *testing.T) {
	z := New()
	z.Add(1, "1")
	z.Add(2, "2")
	z.Add(3, "3")
	z.Limit(1)
	assert(t, z.Count() == 1, "Limit error")
	assert(t, z.Rank("3") == 0, "Limit Rank error")
	z.Add(4.4, "4")
	z.Add(5.5, "5")
	z.Add(0.5, "0.5")
	z.Dump()
	assert(t, z.RevLimit(4) == 0, "RevLimit error")
	assert(t, z.RevLimit(0) == 4, "RevLimit2 error")
}
