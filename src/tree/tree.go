package tree

import (
    "encoding/gob"
)

//
// StringTree
//
// Implements a red-black tree whose search values are ints and whose 
//reference values are strings.
//

type SortedStringListNode struct {
    Str string
    Next *SortedStringListNode
}

func (strNode *SortedStringListNode) index(str string) (int, bool) {
    if str == strNode.Str {
        return 0, true
    } else if strNode.Next != nil {
        depth, ok := strNode.Next.index(str)
        return depth+1, ok
    }
    return 0, false
}

func (strNode *SortedStringListNode) delete(str string) (*SortedStringListNode, bool) {
    if str < strNode.Str {
        return strNode, false
    } else if str > strNode.Str {
        if strNode.Next == nil {
            return strNode, false
        } else {
            var ok bool
            strNode.Next, ok = strNode.Next.delete(str)
            return strNode, ok
        }
    } else {
        return strNode.Next, true
    }
    return strNode, false 
}

func (strNode *SortedStringListNode) insert(str string) (*SortedStringListNode, bool) {
    var ok bool
    if str < strNode.Str {
        newNode := &SortedStringListNode{str,nil}
        newNode.Next = strNode
        return newNode, true
    } else if str > strNode.Str {
        if strNode.Next == nil {
            strNode.Next = &SortedStringListNode{str,nil}
            ok = true
        } else {
            strNode.Next, ok = strNode.Next.insert(str)
        }
    } else {
        return strNode, false
    }
    return strNode, ok 
}

type StringTreeNode struct {
    Left *StringTreeNode
    NodesToLeft int
    Val int
    HeadStr *SortedStringListNode
    Count int
    Right *StringTreeNode
    NodesToRight int
}

type StringTree struct {
    Root *StringTreeNode
}

func (tree *StringTreeNode) insert(val int, str string) (bool) {
    if val < tree.Val {
        tree.NodesToLeft += 1
        if (tree.Left == nil) {
            tree.Left = &StringTreeNode{nil, 0, val, &SortedStringListNode{str,nil}, 1, nil, 0}
            return true
        } else {
            return tree.Left.insert(val, str)
        }
    } else if val > tree.Val {
        tree.NodesToRight += 1
        if (tree.Right == nil) {
            tree.Right = &StringTreeNode{nil, 0, val, &SortedStringListNode{str,nil}, 1, nil, 0}
            return true
        } else {
            return tree.Right.insert(val, str)
        }
    } else {
        var ok bool
        tree.HeadStr, ok = tree.HeadStr.insert(str)
        if ok {
            tree.Count += 1
        }
        return ok
    }
    return false
}

func (tree *StringTree) Insert(val int, str string) (bool) {
    if tree.Root != nil {
        return tree.Root.insert(val,str)
    } else {
        tree.Root = &StringTreeNode{nil, 0, val, &SortedStringListNode{str,nil}, 1, nil, 0}
        return true
    }
    return false
}

func (tree *StringTreeNode) delete(val int, str string) (*StringTreeNode, bool) {
    ok := false
    if val == tree.Val {
        var deleted bool
        tree.HeadStr, deleted = tree.HeadStr.delete(str)
        if deleted {
            tree.Count -= 1
            if tree.HeadStr == nil {
                if tree.Left == nil && tree.Right == nil {
                    return nil, true
                } else if tree.Left == nil && tree.Right != nil {
                    return tree.Right, true
                } else if tree.Left != nil && tree.Right == nil {
                    return tree.Left, true
                } else {
                    panic("This case does not work")
                    tree.Val = tree.Left.Val
                    tree.Left.delete(tree.Left.Val, str)
                    return tree, true
                }
            } else {
                return tree, true
            }
        } else {
            return tree, false
        }
    } else if val < tree.Val {
        if tree.Left != nil {
            tree.Left, ok = tree.Left.delete(val, str)
            if ok {
                tree.NodesToLeft -= 1
            }
        }
    } else if val > tree.Val {
        if tree.Right != nil {
            tree.Right, ok = tree.Right.delete(val, str)
            if ok {
                tree.NodesToRight -= 1
            }
        }
    }
    return tree, ok
}

func (tree *StringTree) Delete(val int, str string) {
    if tree.Root != nil {
        tree.Root, _ = tree.Root.delete(val, str)
    }
}

func (tree *StringTreeNode) rank(val int, str string, lastRank int, stepDir bool) (int, bool) {
    if (stepDir) {
        // left child of parent
        lastRank -= tree.NodesToRight+tree.Count
    } else {
        // right child of parent
        lastRank += tree.NodesToLeft
    }
    if tree.Val == val {
        depth, ok := tree.HeadStr.index(str)
        return lastRank+depth, ok
    } else if val < tree.Val {
        if (tree.Left != nil) {
            return tree.Left.rank(val, str, lastRank, true)
        } else {
            return 0, false
        }
    } else if val > tree.Val {
        if (tree.Right != nil) {
            return tree.Right.rank(val, str, lastRank+tree.Count, false)
        } else {
            return 0, false
        }
    }
    return 0, false
}

func (tree *StringTree) Rank(val int, str string) (int, bool) {
    if tree.Root != nil {
        return tree.Root.rank(val, str, 0, false) 
    } else {
        return 0, false
    }
    return 0, false
}

type ScorePair struct {
    Val int
    Str string
}

func init() {
    gob.Register(SortedStringListNode{})
    gob.Register(StringTreeNode{})
    gob.Register(StringTree{})
}
