package main

import (
	//"fmt"
	"testing"
)

func TestSingleHash(t *testing.T) {
	data := 0
	hash := "4108050209~502633748"
	in := make(chan interface{})
	out := make(chan interface{})
	go SingleHash(in, out)
	in <- data
	result := <-out
	if result.(string) != hash {
		t.Error("Wrong hash: ", result)
	}
}

func TestMultiHash(t *testing.T) {
	data := "4108050209~502633748"
	hash := "29568666068035183841425683795340791879727309630931025356555"
	in := make(chan interface{})
	out := make(chan interface{})
	go MultiHash(in, out)
	in <- data
	result := <-out
	if result.(string) != hash {
		t.Error("Wrong hash: ", result)
	}
}

func TestCombineResults(t *testing.T) {
	data1 := "29568666068035183841425683795340791879727309630931025356555"
	data2 := "4958044192186797981418233587017209679042592862002427381542"
	hash := "29568666068035183841425683795340791879727309630931025356555_4958044192186797981418233587017209679042592862002427381542"
	in := make(chan interface{})
	out := make(chan interface{})
	go CombineResults(in, out)
	in <- data1
	in <- data2
	close(in)
	result := <-out
	if result.(string) != hash {
		t.Error("Wrong hash: ", result)
	}
}
