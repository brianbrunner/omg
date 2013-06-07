package zset

import (
	"encoding/gob"
	"llrb"
)

type ZSet struct {
	Tree *llrb.LLRB
	Set  map[string]int
}

type ScoreMemberPair struct {
	Score  int
	Member string
}

func (smp1 ScoreMemberPair) Less(than llrb.Item) bool {
	smp2 := than.(ScoreMemberPair)
	if smp1.Score < smp2.Score {
		return true
	} else if smp1.Score == smp2.Score &&
		smp1.Member < smp2.Member {
		return true
	} else {
		return false
	}
}

func (zs *ZSet) Insert(val int, key string) {
	if old_val, ok := zs.Set[key]; ok {
		zs.Tree.Delete(ScoreMemberPair{old_val, key})
	}
	zs.Set[key] = val
	zs.Tree.InsertNoReplace(ScoreMemberPair{val, key})
}

func (zs *ZSet) Delete(key string) {
	val := zs.Set[key]
	zs.Tree.Delete(ScoreMemberPair{val, key})
	delete(zs.Set, key)
}

func (zs *ZSet) Card() int {
	return len(zs.Set)
}

func (zs *ZSet) Score(str string) (int, bool) {
	val, ok := zs.Set[str]
	return val, ok
}

//func (zs *ZSet) Rank(str string) (int, bool) {
//    val, ok := zs.Set[str]
//    if ok {
//        return zs.Tree.Rank(val, str)
//    }
//    return 0, false
//}

func New() *ZSet {
	tree := llrb.New()
	set := make(map[string]int)
	zs := &ZSet{tree, set}
	return zs
}

func init() {
	gob.Register(&ZSet{})
	gob.Register(&llrb.LLRB{})
	gob.Register(&llrb.Node{})
	gob.Register(ScoreMemberPair{})
}
