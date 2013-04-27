package tree

type StringTreeNode struct {
    left *StringTreeNode
    val int
    str string
    right *StringTreeNode
}

type StringTree struct {
    root *StringTreeNode
}

func (tree *StringTreeNode) insert(val int, str string) {
    if val < tree.val {
        if (tree.left == nil) {
            tree.left = &StringTreeNode{nil, val, str, nil}
        } else {
            tree.left.insert(val, str)
        }
    } else if val > tree.val {
        if (tree.right == nil) {
            tree.right = &StringTreeNode{nil, val, str, nil}
        } else {
            tree.right.insert(val, str)
        }
    } else {
        tree.str = str
    }
}

func (tree *StringTree) Insert(val int, str string) {
    if tree.root != nil {
        tree.root.insert(val,str)
    }
}

func (tree *StringTreeNode) stringForValue(val int) (string, bool) {
    if val < tree.val {
        if (tree.left == nil) {
            return "", false
        } else {
            return tree.left.stringForValue(val)
        }
    } else if val > tree.val {
        if (tree.right == nil) {
            return "", false
        } else {
            return tree.right.stringForValue(val)
        }
    } else {
        return tree.str, true
    }
    return "", false
}

func (tree *StringTree) StringForValue(val int) (string, bool) {
    if tree.root != nil {
        return tree.root.stringForValue(val)
    }
    return "", false
}
