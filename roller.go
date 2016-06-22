package main

import (
	"flag"
	"log"
	"net"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	channels = NewBroadcastChannels()
)

func ConfigLog(fileName string, maxSizeinMB int, maxBackup int, maxAge int) {
	log.SetOutput(&lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    maxSizeinMB, // megabytes
		MaxBackups: maxBackup,
		MaxAge:     maxAge, //days
	})
}

func IsDirectory(path string) bool {
	if fileInfo, err := os.Stat(path); err != nil {
		return false
	} else {
		return fileInfo.IsDir()
	}
}

func main() {
	logDir := flag.String("d", "stdout", "log dir")
	bindAddress := flag.String("l", "127.0.0.1:6380", "Listen address")
	flag.Parse()
	if *logDir != "stdout" {
		if IsDirectory(*logDir) {
			ConfigLog(*logDir+"/applog.log", 20, 20, 30)
		} else {
			if err := os.Mkdir(*logDir, 0755); err == nil {
				ConfigLog(*logDir+"/applog.log", 20, 20, 30)
			}
		}
	}

	listener, err := net.Listen("tcp", *bindAddress)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error on accept: ", err)
			continue
		}
		client := NewClient(conn, channels)
		go client.Run()
	}
}
