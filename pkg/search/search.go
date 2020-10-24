package search

import (
	"context"
	"io/ioutil"
	"log"
	"strings"
	"sync"
)

type Result struct {
	Phrase  string
	Line    string
	LineNum int64
	ColNum  int64
}

func All(ctx context.Context, phrase string, files []string) <-chan []Result {
	resultChan := make(chan []Result)
	wg := sync.WaitGroup{}

	ctx, cancel := context.WithCancel(ctx)
	for i := 0; i < len(files); i++ {
		wg.Add(1)
		go func(ctx context.Context, filename string, i int, resultChan chan<- []Result) {
			defer wg.Done()
			results := FindAllMatchTextInFile(phrase, filename)
			if len(results) > 0 {
				resultChan <- results
			}
		}(ctx, files[i], i, resultChan)
	}

	go func() {
		defer close(resultChan)
		wg.Wait()

	}()

	cancel()
	return resultChan
}

func Any(ctx context.Context, phrase string, files []string) <-chan Result {
	resultChan := make(chan Result)
	wg := sync.WaitGroup{}

	ctx, cancel := context.WithCancel(ctx)
	for i := 0; i < len(files); i++ {
		wg.Add(1)
		go func(ctx context.Context, filename string, i int, ch chan<- Result) {
			defer wg.Done()
			log.Println("start: ", i, " filename: ", filename)
			select {
			case a := <-ctx.Done():
				log.Println("received cancel: ", i, " a: ", a)
				wg.Done()
				//return
			default:
				result := FindAnyMatchTextInFile(phrase, filename)
				ch <- result
				log.Println("result: ", result)
				if result != (Result{}) {
					//wg.Done()
					log.Println("isEqual: ", result != (Result{}))
					return
				}
			}
		}(ctx, files[i], i, resultChan)
	}

	re := <-resultChan
	cancel()
	log.Println("resultChan in main: ", re)

	go func() {
		defer close(resultChan)
		wg.Wait()
		cancel()

	}()
	return resultChan
}

func FindAllMatchTextInFile(phrase, fileName string) (results []Result) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Println("error while open file: ", err)
		return results
	}

	split := strings.Split(string(data), "\n")
	for i, line := range split {
		if strings.Contains(line, phrase) {
			result := Result{
				Phrase:  phrase,
				Line:    line,
				LineNum: int64(i + 1),
				ColNum:  int64(strings.Index(line, phrase)) + 1,
			}
			results = append(results, result)
		}
	}

	return results
}

func FindAnyMatchTextInFile(phrase, fileName string) (result Result) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Println("error while open file: ", err)
		return result
	}

	split := strings.Split(string(data), "\n")
	for i, line := range split {
		log.Println("line: ", line)
		log.Println("phrase: ", phrase)

		if strings.Contains(line, phrase) {
			return Result{
				Phrase:  phrase,
				Line:    line,
				LineNum: int64(i + 1),
				ColNum:  int64(strings.Index(line, phrase)) + 1,
			}

		}
	}

	return result
}
