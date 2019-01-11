package abtest

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

type ABTest struct {
	NumUsers         int
	WaitMilliseconds int
	DurationSeconds  int
	TestURL          url.URL
}

func defaulter(a ABTest) ABTest {
	if a.NumUsers == 0 {
		a.NumUsers = 10
	}

	if a.WaitMilliseconds == 0 {
		a.WaitMilliseconds = 500
	}

	if a.DurationSeconds == 0 {
		a.DurationSeconds = 30
	}

	return a
}

type TestResult struct {
	WorkerID  int
	TestURL   url.URL
	TimeTaken time.Duration
	Response  string
	Succeeded bool
	RunTime   time.Time
}

func worker(id int, a ABTest, quit chan bool, chResult chan TestResult) {
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: cookieJar}

	for {
		select {
		case <-quit:
			//fmt.Println("Quiting worker")
			return
		default:
			r, err := client.Get(a.TestURL.String())
			if err != nil {
				//todo: stop after some amount of errors?
				log.Printf("Worker %v had error on %v.  error: %v\n", id, a.TestURL.String(), err.Error())
				chResult <- TestResult{WorkerID: id, TestURL: a.TestURL, Response: err.Error(), Succeeded: false, RunTime: time.Now()}
				continue
			} else {
				b, err := ioutil.ReadAll(r.Body)
				if err != nil {
					log.Printf("Worker %v had error on %v.  error: %v\n", id, a.TestURL.String(), err.Error())
					chResult <- TestResult{WorkerID: id, TestURL: a.TestURL, Response: err.Error(), Succeeded: false, RunTime: time.Now()}
					continue
				}

				// fmt.Printf("Worker %v completed a request\n", id)
				chResult <- TestResult{WorkerID: id, TestURL: a.TestURL, Response: string(b), Succeeded: true, RunTime: time.Now()}

			}

		}
		time.Sleep(time.Duration(a.WaitMilliseconds) * time.Millisecond)
	}

}

func StartABtest(a ABTest, results chan TestResult, exit chan bool) []TestResult {
	a = defaulter(a)
	quit := make(chan bool, a.NumUsers)
	//results := make(chan testResult)
	ret := make([]TestResult, 0)

	end := time.After(time.Duration(a.DurationSeconds) * time.Second)

	for i := 0; i < a.NumUsers; i++ {
		go worker(i, a, quit, results)
	}

forLoop:
	for {
		select {
		case t := <-results:
			ret = append(ret, t)
		case <-end:
			for i := 0; i < a.NumUsers; i++ {
				quit <- true
			}

			break forLoop
		}
	}
	exit <- true
	return ret
}
