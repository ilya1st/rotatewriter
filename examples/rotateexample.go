package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ilya1st/rotatewriter"
	"github.com/rs/zerolog"
)

func main() {
	// this is for test
	writer, err := rotatewriter.NewRotateWriter("./logs/test.log", 8)
	if err != nil {
		panic(err)
	}
	sighupChan := make(chan os.Signal, 1)
	signal.Notify(sighupChan, syscall.SIGHUP)
	go func() {
		for {
			_, ok := <-sighupChan
			if !ok {
				return
			}
			fmt.Println("Log rotation")
			writer.Rotate(nil)
		}
	}()
	logger := zerolog.New(writer).With().Timestamp().Logger()
	fmt.Println("Just run in another console and look into logs directory:\n$ killall -HUP rotateexample")
	for {
		logger.Info().Msg("test")
		time.Sleep(500 * time.Millisecond)
	}
}
