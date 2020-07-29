package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func pipelineWorker(in, out chan interface{}, job job, wg *sync.WaitGroup) {

	defer wg.Done()
	defer close(out)
	job(in, out)

}

func ExecutePipeline(freeFlowJobs ...job) {

	wg := &sync.WaitGroup{}

	in := make(chan interface{})

	for _, job := range freeFlowJobs {
		wg.Add(1)
		out := make(chan interface{})
		go pipelineWorker(in, out, job, wg)
		in = out
	}

	wg.Wait()

}

func singleHashWorker(in interface{}, out chan interface{}, wg *sync.WaitGroup, mu *sync.Mutex) {

	defer wg.Done()

	data := fmt.Sprintf("%v", in)

	mu.Lock()
	md5 := DataSignerMd5(data)
	mu.Unlock()

	crc32Chan := make(chan string)
	defer close(crc32Chan)

	go func(data string, out chan string) {
		out <- DataSignerCrc32(data)
	}(data, crc32Chan)

	crc32Md5 := DataSignerCrc32(md5)

	crc32 := <-crc32Chan

	out <- crc32 + "~" + crc32Md5

}

func SingleHash(in, out chan interface{}) {

	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for i := range in {
		wg.Add(1)
		go singleHashWorker(i, out, wg, mu)
	}

	wg.Wait()

}

func multiHashWorker(in interface{}, out chan interface{}, wg *sync.WaitGroup) {

	defer wg.Done()

	resList := make([]string, 6)

	wgCrc32 := &sync.WaitGroup{}
	muResList := &sync.Mutex{}

	for i := 0; i < 6; i++ {

		data := strconv.Itoa(i) + fmt.Sprintf("%v", in)

		wgCrc32.Add(1)

		go func(data string, idx int, resList []string, wg *sync.WaitGroup, mu *sync.Mutex) {
			defer wg.Done()
			data = DataSignerCrc32(data)
			mu.Lock()
			resList[idx] = data
			mu.Unlock()
		}(data, i, resList, wgCrc32, muResList)

	}

	wgCrc32.Wait()

	out <- strings.Join(resList, "")

}

func MultiHash(in, out chan interface{}) {

	wg := &sync.WaitGroup{}

	for i := range in {
		wg.Add(1)
		go multiHashWorker(i, out, wg)
	}

	wg.Wait()

}

func CombineResults(in, out chan interface{}) {

	var resList []string
	for res := range in {
		resList = append(resList, fmt.Sprintf("%v", res))
	}
	sort.Strings(resList)

	out <- strings.Join(resList, "_")
}
