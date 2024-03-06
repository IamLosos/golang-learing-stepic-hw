package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// сюда писать код
const (
	MultiHashCount     = 6
	ChannelsBufferSize = 10
)

func GetSingleHashAsync(s string, wg *sync.WaitGroup, mu *sync.Mutex, out chan<- interface{}) {
	go func() {
		//fmt.Printf("GetSingleHashAsync: start for %s\r\n", s)
		defer wg.Done()

		rch1 := make(chan string)
		go func() {
			rch1 <- DataSignerCrc32(s)
			close(rch1)
		}()

		rch2 := make(chan string)
		go func() {
			mu.Lock()
			md5 := DataSignerMd5(s)
			mu.Unlock()
			rch2 <- DataSignerCrc32(md5)
			close(rch2)
		}()

		r := fmt.Sprintf("%s~%s", <-rch1, <-rch2)
		//fmt.Printf("SingleHash: %s\r\n", r)
		out <- r

		// fmt.Printf("GetSingleHashAsync: finish for %s\r\n", s)
	}()
}

// считает значение crc32(data)+"~"+crc32(md5(data))
func SingleHash(in, out chan interface{}) {
	// fmt.Printf("SingleHash: start; in - %v, out - %v\r\n", in, out)

	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for dataRaw := range in {
		data, ok := dataRaw.(int)
		if !ok {
			panic("SingleHash: cant convert result data to int")
		}

		wg.Add(1)
		GetSingleHashAsync(strconv.Itoa(data), wg, mu, out)
	}

	wg.Wait()
	// fmt.Printf("SingleHash: finish\r\n")
}

type thDataResult struct {
	Key   int
	Value string
}

func GetMultiHashAsync(s string, wg *sync.WaitGroup, out chan<- interface{}) {
	go func() {
		defer wg.Done()

		wgTh := &sync.WaitGroup{}
		// var crs [MultiHashCount]string
		crs := make([]string, MultiHashCount)
		rch1 := make(chan thDataResult)
		for i := 0; i < MultiHashCount; i++ {
			wgTh.Add(1)
			//crs[i] = DataSignerCrc32(fmt.Sprintf("%d%s", i, data))
			go func(i int) {
				defer wgTh.Done()
				rch1 <- thDataResult{
					Key:   i,
					Value: DataSignerCrc32(fmt.Sprintf("%d%s", i, s)),
				}
			}(i)
		}

		go func() { //async because channel is unbuffered
			wgTh.Wait()
			close(rch1) // Close the channel when all workers have finished
		}()

		for r := range rch1 {
			crs[r.Key] = r.Value
		}

		r := strings.Join([]string(crs), "")
		// fmt.Printf("MultiHash: %s\r\n", r)
		out <- r
	}()
}

// считает значение crc32(th+data)) (конкатенация цифры, приведённой к строке и строки), где th=0..5
func MultiHash(in, out chan interface{}) {
	// fmt.Printf("MultiHash: start; in - %v, out - %v\r\n", in, out)

	wg := &sync.WaitGroup{}

	for dataRaw := range in {
		data, ok := dataRaw.(string)
		if !ok {
			panic("MultiHash: cant convert result data to string")
		}

		wg.Add(1)
		GetMultiHashAsync(data, wg, out)
	}

	wg.Wait()
	// fmt.Printf("MultiHash: finish\r\n")
}

func SortStringSliceAscend(strings []string) {
	sort.Slice(strings, func(i, j int) bool {
		return strings[i] < strings[j]
	})
}

func CombineResults(in, out chan interface{}) {
	// fmt.Printf("CombineResults: start; in - %v, out - %v\r\n", in, out)

	i := 0
	crs := make([]string, 0)
	for dataRaw := range in {
		data, ok := dataRaw.(string)
		if !ok {
			panic("CombineResults: cant convert result data to string")
		}

		//fmt.Printf("CombineResults: %s\r\n", data)
		crs = append(crs, data)
		i++
	}

	SortStringSliceAscend(crs)
	fmt.Println(crs)
	r := strings.Join([]string(crs), "_")
	// fmt.Printf("CombineResults: %s\r\n", r)
	out <- r
	// fmt.Printf("CombineResults: finish\r\n")
}

func ExecutePipeline(freeFlowJobs ...job) {
	wg := &sync.WaitGroup{}
	wg.Add(len(freeFlowJobs))

	in := make(chan interface{}, ChannelsBufferSize)
	for idx, j := range freeFlowJobs {
		out := make(chan interface{}, ChannelsBufferSize)
		go func(j job, idx int, inParam chan interface{}) {
			defer wg.Done()
			j(inParam, out)
			close(out)
		}(j, idx, in)

		in = out
	}

	wg.Wait()
}

func RunPipeline() {
	fmt.Println("RunPipeline: start")

	inputData := []int{0, 1, 1, 2, 3, 5, 8}
	//inputData := []int{0, 1}

	testResult := "NOT_SET"

	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, n := range inputData {
				out <- n
			}
		}),
		//
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		//
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, ok := dataRaw.(string)
			if !ok {
				panic("RunPipeline: cant convert result data to string")
			}
			testResult = data
			fmt.Print(testResult)
		}),
	}

	start := time.Now()

	ExecutePipeline(hashSignJobs...)

	end := time.Since(start)
	fmt.Printf("RunPipeline complited in %s\r\n", end)
}

func main() {
	fmt.Println("main: start")

	RunPipeline()
}
