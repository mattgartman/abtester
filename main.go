package main

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	tm "github.com/buger/goterm"
	"github.com/mattgartman/abtester/abtest"
)

func main() {
	x, _ := url.Parse("http://localhost:8080")
	matchA := "Hello World A"
	test := abtest.ABTest{TestURL: *x, DurationSeconds: 60, NumUsers: 30}
	chResults := make(chan abtest.TestResult)
	chExit := make(chan bool)
	var ret []abtest.TestResult
	go abtest.StartABtest(test, chResults, chExit)
	var retMap = make(map[int][]abtest.TestResult)

forLoop:
	for {
		select {
		case t := <-chResults:
			ret = append(ret, t)
			retMap[t.WorkerID] = append(retMap[t.WorkerID], t)
			var keys []int
			for k := range retMap {
				keys = append(keys, k)
			}
			sort.Ints(keys)

			tm.Flush()
			tm.Clear()
			tm.Println(tm.Color(x.String(), tm.RED))
			aCount := 1
			bCount := 1
			for k := range keys {
				tm.Printf("%v |", k)
				for _, v := range retMap[k] {
					val := strings.TrimSuffix(v.Response, "\n")
					if strings.Contains(val, matchA) {
						tm.Print("A")
						aCount++
					} else {
						tm.Print("B")
						bCount++
					}
				}
				tm.Println("")
			}
			tm.Println(tm.Color(fmt.Sprintf("A: %v (%2.0f) B: %v (%2.0f) Total: %v", aCount, float64(aCount)/float64(aCount+bCount)*100, bCount, float64(bCount)/float64(aCount+bCount)*100, aCount+bCount), tm.BLUE))

		case <-chExit:
			break forLoop
		}
	}

}
