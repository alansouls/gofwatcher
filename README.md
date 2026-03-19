# Go File Watcher

## Overview

This is simple file watcher lib written in go with an exposed C-ABI so you can easily integrate with any other project in any other language. The steps bellow will provide a quick guide on how to build this for the platform you desire and how to use it in the c++ example provided.

## Building
To build, just run the command:

```
go build -o out/gofwatcher.a -buildmode=c-archive gofwatcher.go
```

This will generate the gofwatcher.h and gofwatcher.a files in the out directory, if you're on Windows you might want to change the gofwatcher.a to be gofwatcher.lib. 
You can also choose to generate a dynamic linked library by changing -buildmode=c-archive to -build-mode=c-shared

## Running the example
To run the example, you need a c++ compiler with c++23 support, if you're using g++ simply do: 

```
g++ example.cpp -std=c++23 -o out/example ../lib/osx_arm/gofwatcher.a -lpthread
```

The example simply watches for files under a given path and an optional file regex so you only observe files that the names match the regex expression. Usage example: 
```
./example "../../test" ".*\.cpp$"
```

This will watch for any .cpp file created, modified or deleted inside the test folder in the root of this repo
