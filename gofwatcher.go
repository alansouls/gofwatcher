package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

func watch(waitGroup *sync.WaitGroup, path string) {
	defer waitGroup.Done()

	for {
		entries, err := os.ReadDir(path)

		if err != nil {
			log.Fatal(err)
		}

		for len(entries) > 0 {
			tempEntries := entries
			entries = nil
			for _, entry := range tempEntries {
				if entry.IsDir() {
					entries = append(entries, entry)
				} else {
					fmt.Println(entry.Name())
				}
			}
		}

		time.Sleep(1 * time.Second)
	}

}

func main() {
	var wg sync.WaitGroup

	dir := os.Args[1]

	wg.Add(1)
	go watch(&wg, dir)

	wg.Wait()
}
