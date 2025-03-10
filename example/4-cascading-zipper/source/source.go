package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/yomorun/yomo"
)

func main() {
	// connect to YoMo-Zipper.
	source := yomo.NewSource(
		"yomo-source",
		"localhost:9001",
		yomo.WithCredential("token:z1"),
	)
	err := source.Connect()
	if err != nil {
		log.Printf("[source] ❌ Emit the data to YoMo-Zipper failure with err: %v", err)
		return
	}
	defer source.Close()

	// generate mock data and send it to YoMo-Zipper.
	err = generateAndSendData(source)
	if err != nil {
		log.Printf("[source] >>>> ERR >>>> %v", err)
		os.Exit(1)
	}
	select {}
}

func generateAndSendData(stream yomo.Source) error {
	for {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		data := []byte(fmt.Sprintf("%d", rnd.Uint32()))
		// send data via QUIC stream.
		// broadcast this message to cascading zippers using `Broadcast` method
		err := stream.Broadcast(0x33, data)
		if err != nil {
			log.Printf("[source] ❌ Emit %v to YoMo-Zipper failure with err: %v", data, err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		log.Printf("[source] ✅ Emit %s to YoMo-Zipper", data)
		time.Sleep(1000 * time.Millisecond)
	}
}
