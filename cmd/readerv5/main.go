package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/juruen/rmapi/encoding/rm"
	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func _main() error {
	filename := os.Args[1]
	file, err := os.Open(filename)

	if err != nil {
		return err
	}
	defer file.Close()

	pageData, err := ioutil.ReadAll(file)

	if err != nil {
		log.Fatal("cant read fil")
		return err
	}
	rm := rm.New()
	err = rm.UnmarshalBinary(pageData)
	if err != nil {
		return err
	}
	fmt.Printf("Number of Layer: %d\n", len(rm.Layers))
	for i, layer := range rm.Layers {
		fmt.Printf("Layer: %d num lines:%d\n", i, len(layer.Lines))
		for j, line := range layer.Lines {
			fmt.Printf("\tLine: %d points: %d\n", j, len(line.Points))
			for _, point := range line.Points {
				fmt.Printf("\t\t\tX: %f Y: %f speed: %f width: %f\n", point.X, point.Y, point.Speed, point.Width)
			}
		}
	}

	return nil
}

func main() {
	prefixed := &prefixed.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceFormatting: true,
		ForceColors:     true,
	}
	log.SetFormatter(prefixed)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	err := _main()
	if err != nil {
		log.Fatal(err)
	}
}
