package main

import (
	//	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func main() {
}

func ExecutePipeline(jobs ...job) {
	var ch1, ch2 chan interface{}
	wgJobs := &sync.WaitGroup{}
	ch2 = make(chan interface{})
	for _, f := range jobs {
		in := ch1
		out := ch2
		wgJobs.Add(1)
		go func(wg *sync.WaitGroup, f job, in, out chan interface{}) {
			defer wg.Done()
			//fmt.Println(in, out, f)
			f(in, out)
			close(out)
		}(wgJobs, f, in, out)
		ch1 = ch2
		ch2 = make(chan interface{})
	}
	wgJobs.Wait()
}

func SingleHash(in, out chan interface{}) {
	mutex := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for data := range in {
		parts := make([]string, 2)
		value, ok := data.(string)
		if !ok {
			value = strconv.Itoa(data.(int))
		}
		wg.Add(2)
		go func(wg *sync.WaitGroup, slot int, mutex *sync.Mutex) {
			defer wg.Done()
			crc := DataSignerCrc32(value)
			mutex.Lock()
			parts[slot] = crc
			mutex.Unlock()
		}(wg, 0, mutex)

		go func(wg *sync.WaitGroup, slot int, mutex *sync.Mutex) {
			defer wg.Done()
			crc := DataSignerCrc32(DataSignerMd5(value))
			mutex.Lock()
			parts[slot] = crc
			mutex.Unlock()
		}(wg, 1, mutex)
		wg.Wait()
		result := strings.Join(parts, "~")
		out <- result
	}
}

func MultiHash(in, out chan interface{}) {
	mutex := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for data := range in {
		var result string
		parts := make([]string, 6)
		for i := 0; i < 6; i++ {
			wg.Add(1)
			go func(wg *sync.WaitGroup, i int, parts []string, mutex *sync.Mutex) {
				defer wg.Done()
				crc := DataSignerCrc32(strconv.Itoa(i) + data.(string))
				mutex.Lock()
				parts[i] = crc
				mutex.Unlock()
			}(wg, i, parts, mutex)
		}
		wg.Wait()
		result = strings.Join(parts, "")
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
