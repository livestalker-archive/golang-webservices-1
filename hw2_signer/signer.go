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
	wgGlobal := &sync.WaitGroup{}
	onceMD5 := &sync.Mutex{}
	for data := range in {
		wgGlobal.Add(1)
		go func(data interface{}) {
			defer wgGlobal.Done()
			parts := make([]string, 2)
			mutex := &sync.Mutex{}
			wg := &sync.WaitGroup{}
			value, ok := data.(string)
			if !ok {
				value = strconv.Itoa(data.(int))
			}
			wg.Add(2)
			go func(slot int) {
				defer wg.Done()
				crc := DataSignerCrc32(value)
				mutex.Lock()
				defer mutex.Unlock()
				parts[slot] = crc
			}(0)
			go func(slot int) {
				defer wg.Done()
				onceMD5.Lock()
				md5 := DataSignerMd5(value)
				onceMD5.Unlock()
				crc := DataSignerCrc32(md5)
				mutex.Lock()
				defer mutex.Unlock()
				parts[slot] = crc
			}(1)
			wg.Wait()
			out <- strings.Join(parts, "~")
		}(data)
	}
	wgGlobal.Wait()
}

func SingleHashOld(in, out chan interface{}) {
	wgGlobal := &sync.WaitGroup{}
	onceMD5 := &sync.Mutex{}
	for data := range in {
		wgGlobal.Add(1)
		go func(data interface{}) {
			defer wgGlobal.Done()
			parts := make([]string, 2)
			mutex := &sync.Mutex{}
			wg := &sync.WaitGroup{}
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
				onceMD5.Lock()
				md5 := DataSignerMd5(value)
				onceMD5.Unlock()
				crc := DataSignerCrc32(md5)
				mutex.Lock()
				parts[slot] = crc
				mutex.Unlock()
			}(wg, 1, mutex)
			wg.Wait()
			result := strings.Join(parts, "~")
			out <- result
		}(data)
	}
	wgGlobal.Wait()
}

func MultiHash(in, out chan interface{}) {
	wgGlobal := &sync.WaitGroup{}
	for data := range in {
		wgGlobal.Add(1)
		go func(data interface{}) {
			defer wgGlobal.Done()
			parts := make([]string, 6)
			wg := &sync.WaitGroup{}
			mutex := &sync.Mutex{}
			for i := 0; i < 6; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					crc := DataSignerCrc32(strconv.Itoa(i) + data.(string))
					mutex.Lock()
					parts[i] = crc
					mutex.Unlock()
				}(i)
			}
			wg.Wait()
			out <- strings.Join(parts, "")
		}(data)
	}
	wgGlobal.Wait()
}

func MultiHashOld(in, out chan interface{}) {
	wgGlobal := &sync.WaitGroup{}
	for data := range in {
		wgGlobal.Add(1)
		go func(data interface{}) {
			defer wgGlobal.Done()
			var result string
			mutex := &sync.Mutex{}
			wg := &sync.WaitGroup{}
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
		}(data)
	}
	wgGlobal.Wait()
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
