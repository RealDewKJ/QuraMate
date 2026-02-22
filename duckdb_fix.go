package main

/*
#include <stdlib.h>
#include <string.h>

struct ArrowSchema {
  const char* format;
  const char* name;
  const char* metadata;
  long flags;
  long n_children;
  struct ArrowSchema** children;
  struct ArrowSchema* dictionary;
  void (*release)(struct ArrowSchema*);
  void* private_data;
};

struct ArrowArray {
  long length;
  long null_count;
  long offset;
  long n_buffers;
  long n_children;
  const void** buffers;
  struct ArrowArray** children;
  struct ArrowArray* dictionary;
  void (*release)(struct ArrowArray*);
  void* private_data;
};

struct ArrowArrayStream {
  int (*get_schema)(struct ArrowArrayStream*, struct ArrowSchema*);
  int (*get_next)(struct ArrowArrayStream*, struct ArrowArray*);
  const char* (*get_last_error)(struct ArrowArrayStream*);
  void (*release)(struct ArrowArrayStream*);
  void* private_data;
};

int ArrowSchemaIsReleased(const struct ArrowSchema* schema) {
  return schema->release == NULL;
}

void ArrowSchemaMarkReleased(struct ArrowSchema* schema) {
  schema->release = NULL;
}

void ArrowSchemaRelease(struct ArrowSchema* schema) {
  if (schema->release != NULL) {
    schema->release(schema);
    schema->release = NULL;
  }
}

int ArrowArrayIsReleased(const struct ArrowArray* array) {
  return array->release == NULL;
}

void ArrowArrayMarkReleased(struct ArrowArray* array) {
  array->release = NULL;
}

void ArrowArrayRelease(struct ArrowArray* array) {
  if (array->release != NULL) {
    array->release(array);
    array->release = NULL;
  }
}

void ArrowArrayMove(struct ArrowArray* src, struct ArrowArray* dest) {
  memcpy(dest, src, sizeof(struct ArrowArray));
  src->release = NULL;
}

void ArrowArrayStreamRelease(struct ArrowArrayStream* stream) {
  if (stream->release != NULL) {
    stream->release(stream);
    stream->release = NULL;
  }
}

void ArrowArrayStreamMove(struct ArrowArrayStream* src, struct ArrowArrayStream* dest) {
  memcpy(dest, src, sizeof(struct ArrowArrayStream));
  src->release = NULL;
}
*/
import "C"
