package zset

//#include "skiplist.h"
import "C"

import (
	"reflect"
	"runtime"
	"unsafe"
)

const DEFAULT_TBL_LEN = 16

type zset struct {
	sl  *C.skiplist
	tbl map[string]float64
}

func tocstring(s string) (*C.char, C.size_t) {
	v := (*reflect.StringHeader)(unsafe.Pointer(&s))
	return (*C.char)(unsafe.Pointer(v.Data)), C.size_t(v.Len)
}

func newslobj(s string) *C.slobj {
	p, l := tocstring(s)
	return C.slCreateObj(p, l)
}

func New() *zset {
	z := &zset{C.slCreate(), make(map[string]float64, DEFAULT_TBL_LEN)}
	runtime.SetFinalizer(z, func(z *zset) {
		C.slFree(z.sl)
		z.tbl = nil
	})
	return z
}

func (z *zset) Add(score float64, member string) {
	if old, ok := z.tbl[member]; ok {
		if old == score {
			return
		}
		var obj C.slobj
		obj.ptr, obj.length = tocstring(member)
		C.slDelete(z.sl, C.double(old), &obj)
	}
	C.slInsert(z.sl, C.double(score), newslobj(member))
	z.tbl[member] = score
}

func (z *zset) Rem(member string) {
	if score, ok := z.tbl[member]; ok {
		var obj C.slobj
		obj.ptr, obj.length = tocstring(member)
		C.slDelete(z.sl, C.double(score), &obj)
		delete(z.tbl, member)
	}
}

func (z *zset) Count() int {
	return int(z.sl.length)
}

func (z *zset) Score(member string) (float64, bool) {
	score, ex := z.tbl[member]
	return score, ex
}

func (z *zset) Range(r1, r2 int) []string {
	if r1 < 1 {
		r1 = 1
	}
	if r2 < 1 {
		r2 = 1
	}
	var reverse, rangelen int
	if r1 <= r2 {
		reverse = 0
		rangelen = r2 - r1 + 1
	} else {
		reverse = 1
		rangelen = r1 - r2 + 1
	}
	node := C.slGetNodeByRank(z.sl, C.ulong(r1))
	result := make([]string, 0, rangelen)
	rr := C.int(reverse)
	for n := 0; node != nil && n < rangelen; {
		result = append(result, C.GoStringN(node.obj.ptr, C.int(node.obj.length)))
		node = C.getNextNode(node, rr)
		n++
	}
	return result
}

func (z *zset) reverseRank(r int) int {
	return z.Count() - r + 1
}

func (z *zset) RevRange(r1, r2 int) []string {
	return z.Range(z.reverseRank(r1), z.reverseRank(r2))
}

func (z *zset) RangeByScore(s1, s2 float64) []string {
	var reverse int
	var node *C.skiplistNode
	cs1, cs2 := C.double(s1), C.double(s2)
	if s1 <= s2 {
		reverse = 0
		node = C.slFirstInRange(z.sl, cs1, cs2)
	} else {
		reverse = 1
		node = C.slLastInRange(z.sl, cs2, cs1)
	}

	result := make([]string, 0)
	rr := C.int(reverse)
	for node != nil {
		if reverse == 1 {
			if node.score < cs2 {
				break
			}
		} else {
			if node.score > cs2 {
				break
			}
		}
		result = append(result, C.GoStringN(node.obj.ptr, C.int(node.obj.length)))
		node = C.getNextNode(node, rr)
	}
	return result
}

// rank is 1-based, 0 not found
func (z *zset) Rank(member string) int {
	score, ex := z.tbl[member]
	if !ex {
		return 0
	}
	var obj C.slobj
	obj.ptr, obj.length = tocstring(member)
	rank := C.slGetRank(z.sl, C.double(score), &obj)
	return int(rank)
}

func (z *zset) RevRank(member string) int {
	rank := z.Rank(member)
	if rank != 0 {
		rank = z.reverseRank(rank)
	}
	return rank
}

func (z *zset) deleteByRank(from, to int) int {
	if from > to {
		from, to = to, from
	}
	return int(C.slDeleteByRank(z.sl, C.uint(from), C.uint(to), unsafe.Pointer(z)))
}

//export delCb
func delCb(p unsafe.Pointer, obj *C.slobj) {
	z := (*zset)(p)
	member := C.GoStringN(obj.ptr, C.int(obj.length))
	delete(z.tbl, member)
}

func (z *zset) Limit(count int) int {
	total := z.Count()
	if total <= count {
		return 0
	}
	return z.deleteByRank(count+1, total)
}

func (z *zset) RevLimit(count int) int {
	total := z.Count()
	if total <= count {
		return 0
	}
	from := z.reverseRank(count + 1)
	to := z.reverseRank(total)
	return z.deleteByRank(from, to)
}

func (z *zset) Dump() {
	C.slDump(z.sl)
}
