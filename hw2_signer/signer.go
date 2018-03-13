package main

import (
	//"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
}

func ExecutePipeline(jobs ...job) {
	var ch1, ch2 chan interface{}
	ch2 = make(chan interface{})
	for _, f := range jobs {
		in := ch1
		out := ch2
		//fmt.Println("Run job: ", ix, f, in, out)
		go f(in, out)
		ch1 = ch2
		ch2 = make(chan interface{})
	}
	time.Sleep(10 * time.Millisecond)
}

func SingleHash(in, out chan interface{}) {
	for data := range in {
		value, ok := data.(string)
		if !ok {
			value = strconv.Itoa(data.(int))
		}
		result := DataSignerCrc32(value) + "~" + DataSignerCrc32(DataSignerMd5(value))
		out <- result
	}
}

func MultiHash(in, out chan interface{}) {
	for data := range in {
		var result string
		for i := 0; i < 6; i++ {
			result += DataSignerCrc32(strconv.Itoa(i) + data.(string))
		}
		out <- result
	}
}

func CombineResults(in, out chan interface{}) {
	parts := make([]string, 0, 10)
	for data := range in {
		parts = append(parts, data.(string))
	}
	sort.Strings(parts)
	result := strings.Join(parts, "_")
	out <- result
}
