package main

import (
	"fmt"
	"io"
	"os"

	v6 "github.com/ddvk/reader/v6"
	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func parseSceneFile(file io.ReadSeekCloser) (err error) {
	headerLength := 0x2b
	buffer := make([]byte, headerLength)
	_, err = io.ReadFull(file, buffer)
	if err != nil {
		return err
	}
	//todo: checkversion blah blah

	sceneParser := v6.SceneReader{}
	scene, err := sceneParser.ExtractScene(file)
	if err != nil {
		return
	}
	log.Info("parsed: ", scene)

	fmt.Printf("Number of Layer: %d\n", len(scene.Layers))
	for i, layer := range scene.Layers {
		fmt.Printf("Layer: %d num lines:%d\n", i, len(layer.Lines))
		for j, line := range layer.Lines {
			fmt.Printf("\tLine: %d points: %d\n", j, len(line.Line.Value.Points))
			for _, point := range line.Line.Value.Points {
				fmt.Printf("\t\t\tX: %f Y: %f speed: %d width: %d\n", point.X, point.Y, point.Speed, point.Width)
			}
		}
	}
	return
}

func _main() error {
	if len(os.Args) < 2 {
		log.Print("missing file")
		return nil
	}
	filename := os.Args[1]
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return parseSceneFile(file)
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
