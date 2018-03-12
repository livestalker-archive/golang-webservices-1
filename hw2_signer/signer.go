package main

import (
	"fmt"
)

func main() {
}

func ExecutePipeline(jobs ...job) {
	for ix, f := range jobs {
		fmt.Println(ix, f)
	}
}

func SingleHash(in, out chan interface{}) {
}

func MultiHash(in, out chan interface{}) {
}

func CombineResults(in, out chan interface{}) {
}
