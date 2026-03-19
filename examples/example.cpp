#include "../out/gofwatcher.h"

#include <iostream>
#include <print>

static void callback(FilesChangedMessage *changes, size_t count) {
  for (size_t i = 0; i < count; ++i) {
    auto change = changes[i];
    switch (change.changeType) {
    case 0:
      std::println("File added: {}", change.fileName);
      break;
    case 1:
      std::println("File modified: {}", change.fileName);
      break;
    case 2:
      std::println("File deleted: {}", change.fileName);
      break;
    default:
      std::println("Invalid change type on file {}", change.fileName);
      break;
    }
  }
}

auto main(int argc, char **argv) -> int32_t {
  if (argc < 2) {
    std::println("You need to provide a path");
  }

  char *path = argv[1];

  uintptr_t contextHandle = gofwatcher_beginWatch(path, &callback);

  int stop;
  std::cin >> stop;

  gofwatcher_stopWatch(contextHandle);
}
