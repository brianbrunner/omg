package tracker

import (
    "time"
    "fmt"
)

type Lapper struct {
    Last time.Time
}

func (l *Lapper) Lap(str string) {
    fmt.Println(str,time.Since(l.Last).Nanoseconds())
    l.Last = time.Now()
}
