package zset

import (
    "tree"
    "encoding/gob"
)

type ZSet struct {
    BTree *tree.StringTree
    Set map[string]int
}

func (zs *ZSet) Insert(val int, key string) {
    if val, ok := zs.Set[key]; ok {
        zs.BTree.Delete(val, key)
    }
    zs.Set[key] = val
    zs.BTree.Insert(val, key)
}

func (zs *ZSet) Delete(key string) {
    val := zs.Set[key]
    zs.BTree.Delete(val, key)
    delete(zs.Set, key)
}

func (zs *ZSet) Card() int {
    return len(zs.Set)
}

func (zs *ZSet) Score(str string) (int, bool) {
    val, ok := zs.Set[str]
    return val, ok
}

func (zs *ZSet) Rank(str string) (int, bool) {
    val, ok := zs.Set[str]
    if ok {
        return zs.BTree.Rank(val, str)
    }
    return 0, false
}

func New() ZSet {
    st := &tree.StringTree{}
    mp := make(map[string]int)
    zs := ZSet{st,mp}
    return zs
}

func init() {
    gob.Register(ZSet{})
}
