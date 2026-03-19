package main

/*
#include <stdlib.h>
#include <stdint.h>

typedef struct FilesChangedMessage
{
	char* fileName;
	int32_t changeType;
}

// Define your callback function pointer types here
typedef void (*OnFileChangedCallback)(FilesChangedMessage* files, size_t count);
*/
import "C"
import (
	"fmt"
	"log"
	"maps"
	"os"
	"runtime/cgo"
	"strings"
	"time"
)

type FileWatcherContext struct {
	Path               string
	InteruptChannel    chan bool
	FileChangeCallback C.OnFileChangedCallback
}

func getFilesRecursive(path string) map[string]int64 {
	if path == "" {
		return map[string]int64{}
	}

	path = strings.Trim(path, "/")

	entries, err := os.ReadDir(path)

	if err != nil {
		log.Fatal(err)
	}

	files := map[string]int64{}

	for _, entry := range entries {
		if entry.IsDir() {
			maps.Copy(files, getFilesRecursive(path+"/"+entry.Name()))
		} else {
			fileInfo, err := entry.Info()

			if err != nil {
				log.Fatal(err)
			}
			files[path+"/"+entry.Name()] = fileInfo.ModTime().Unix()
		}
	}

	return files
}

func watch(context *FileWatcherContext) {
	filesMap := map[string]int64{}

	for {
		select {
		case <-context.InteruptChannel:
			return
		default:
			newFilesMap := getFilesRecursive(context.Path)

			changes := make(C.FilesChangedMessage, 0)

			for filePath, newLastMod := range newFilesMap {
				oldLastMod, exists := filesMap[filePath]

				if !exists {
					changes = append(changes, C.FilesChangedMessage{})
				} else if oldLastMod != newLastMod {
					fmt.Println("File modified: {}", filePath)
				}
			}

			for filePath := range filesMap {
				_, exists := newFilesMap[filePath]

				if !exists {
					fmt.Println("File deleted: {}", filePath)
				}
			}

			filesMap = newFilesMap

			time.Sleep(1 * time.Second)
		}
	}
}

// export gofwatcher_beginWatch
func beginWatch(path *C.char, fileChangeCallback C.OnFileChangedCallback) C.uintptr_t {
	context := FileWatcherContext{
		Path:               C.GoString(path),
		InteruptChannel:    make(chan bool),
		FileChangeCallback: fileChangeCallback,
	}

	go watch(&context)
	handle := cgo.NewHandle(&context)
	return C.uintptr_t(handle)
}

// export gofwatcher_stopWatch
func stopWatch(contextHandle C.uintptr_t) {
	goHandle := cgo.Handle(contextHandle)
	defer goHandle.Delete()
	context := goHandle.Value().(*FileWatcherContext)
	context.InteruptChannel <- true
}

func main() {

}
