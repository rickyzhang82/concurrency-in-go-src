package main

import (
	"fmt"
	"net/http"
)

func isChannelClosed(done <-chan interface{}) bool {
	var isClosed bool
	select {
	case _, ok := <-done:
		if ok {
			isClosed = false
		} else {
			isClosed = true
		}
	default:
		isClosed = false
	}
	return isClosed
}

func main() {
	type Result struct { // <1>
		Error    error
		Response *http.Response
	}
	checkStatus := func(done <-chan interface{}, urls ...string) <-chan Result { // <2>
		results := make(chan Result)
		go func() {
			defer func() {
				close(results)
				fmt.Println("Close results")
			}()

			for _, url := range urls {
				var result Result
				resp, err := http.Get(url)
				result = Result{Error: err, Response: resp} // <3>
				select {
				case <-done:
					fmt.Println("checkStatus goroutine: get close done!")
					return
				case results <- result: // <4>
				}
			}
		}()
		return results
	}
	done := make(chan interface{})

	errCount := 0
	urls := []string{"a", "https://www.google.com", "b", "c", "d"}
	for result := range checkStatus(done, urls...) {
		if result.Error != nil {
			fmt.Printf("error: %v\n", result.Error)
			errCount++
			if errCount >= 3 {
				fmt.Println("Too many errors, breaking!")
				close(done)
				break
			}
			continue
		}
		fmt.Printf("Response: %v\n", result.Response.Status)
	}
	if !isChannelClosed(done) {
		fmt.Println("Force close(done) in the end.")
		close(done)
	} else {
		fmt.Println("No need to force close(done) in the end.")
	}
}
