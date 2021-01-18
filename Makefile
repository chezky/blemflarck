UNAME := $(shell uname)

ifeq ($(UNAME), Linux)
TARGET = blem
else
TARGET = blem.exe
endif

build:
	go build -o $(TARGET) main.go

blem:
	go build -o $(TARGET) main.go
