package main

/*
#include <stdlib.h>
#include <stdint.h>
#include <string.h>

typedef struct
{
	char* fileName;
	int32_t changeType;
} FilesChangedMessage ;

typedef void (*OnFileChangedCallback)(FilesChangedMessage* changes, size_t count);

static void gofwatcher_invoke_callback(OnFileChangedCallback callback, FilesChangedMessage* changes, size_t count) {
	callback(changes, count);
}
*/
import "C"

import (
	"log"
	"maps"
	"os"
	"regexp"
	"runtime/cgo"
	"strings"
	"time"
	"unsafe"
)

type FileWatcherContext struct {
	Path               string
	FileRegex          *regexp.Regexp
	InteruptChannel    chan bool
	FileChangeCallback C.OnFileChangedCallback
}

func getFilesRecursive(path string, fileRegex *regexp.Regexp) map[string]int64 {
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
			maps.Copy(files, getFilesRecursive(path+"/"+entry.Name(), fileRegex))
		} else {
			fileInfo, err := entry.Info()

			if err != nil {
				log.Fatal(err)
			}

			if fileRegex == nil || fileRegex.MatchString(entry.Name()) {
				files[path+"/"+entry.Name()] = fileInfo.ModTime().Unix()
			}
		}
	}

	return files
}

func createFilesChangedMessage(fileName string, changeType int32) C.FilesChangedMessage {
	return C.FilesChangedMessage{
		fileName:   C.CString(fileName),
		changeType: C.int32_t(changeType),
	}
}

func freeFilesChangedMessage(message *C.FilesChangedMessage) {
	C.free(unsafe.Pointer(message.fileName))
}

func sliceToCArray(fileMessageSlice []C.FilesChangedMessage) unsafe.Pointer {
	cArray := C.malloc(C.size_t(len(fileMessageSlice)) * C.size_t(unsafe.Sizeof(C.FilesChangedMessage{})))
	// convert the C array to a Go Array so we can index it
	tempSlice := (*[1<<30 - 1]C.FilesChangedMessage)(cArray)
	copy(tempSlice[:], fileMessageSlice)
	return cArray
}

func watch(context *FileWatcherContext) {
	filesMap := map[string]int64{}

	for {
		select {
		case <-context.InteruptChannel:
			return
		default:
			newFilesMap := getFilesRecursive(context.Path, context.FileRegex)

			changes := make([]C.FilesChangedMessage, 0)

			for filePath, newLastMod := range newFilesMap {
				oldLastMod, exists := filesMap[filePath]

				if !exists {
					changes = append(changes, createFilesChangedMessage(filePath, 0))
				} else if oldLastMod != newLastMod {
					changes = append(changes, createFilesChangedMessage(filePath, 1))
				}
			}

			for filePath := range filesMap {
				_, exists := newFilesMap[filePath]

				if !exists {
					changes = append(changes, createFilesChangedMessage(filePath, 2))
				}
			}

			if len(changes) > 0 {
				cArray := sliceToCArray(changes)
				C.gofwatcher_invoke_callback(context.FileChangeCallback, (*C.FilesChangedMessage)(cArray), C.size_t(len(changes)))
				C.free(cArray)
			}

			filesMap = newFilesMap

			time.Sleep(1 * time.Second)
		}
	}
}

//export gofwatcher_beginWatch
func gofwatcher_beginWatch(path *C.char, fileChangeCallback C.OnFileChangedCallback, fileRegex *C.char) C.uintptr_t {

	goFileRegexp := (*regexp.Regexp)(nil)

	if fileRegex != nil {
		innerGoFileRegex, err := regexp.Compile(C.GoString(fileRegex))

		if err != nil {
			log.Fatal(err)
		}

		goFileRegexp = innerGoFileRegex
	}

	context := FileWatcherContext{
		Path:               C.GoString(path),
		FileRegex:          goFileRegexp,
		InteruptChannel:    make(chan bool),
		FileChangeCallback: fileChangeCallback,
	}

	go watch(&context)
	handle := cgo.NewHandle(&context)
	return C.uintptr_t(handle)
}

//export gofwatcher_stopWatch
func gofwatcher_stopWatch(contextHandle C.uintptr_t) {
	goHandle := cgo.Handle(contextHandle)
	defer goHandle.Delete()
	context := goHandle.Value().(*FileWatcherContext)
	context.InteruptChannel <- true
}

func main() {

}
