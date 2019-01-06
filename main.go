package main

import (
	"net/url"
	"sort"

	tm "github.com/buger/goterm"
	"github.com/mattgartman/gophercises/abtest/abtester"
)

func main() {
	x, _ := url.Parse("http://localhost:8080")
	test := abtester.ABTest{TestURL: *x, DurationSeconds: 12}
	a := abtester.StartABtest(test)

	sort.Slice(a, func(i int, j int) bool {
		if a[i].WorkerID == a[j].WorkerID {
			return a[i].RunTime.Before(a[j].RunTime)
		}
		return a[i].WorkerID < a[j].WorkerID
	})
	tm.Clear()
	for _, l := range a {
		tm.Printf("%v %v %v", l.WorkerID, l.RunTime.Format("15:04:05"), l.Response)
		tm.Flush()
	}
}
