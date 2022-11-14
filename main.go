package main

import (
	"io"
	v6 "myreader/v6"
	"os"

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
	for _, la := range scene.Layers {
		log.Infof("\t %v", la)
		for _, li := range la.Lines {
			log.Infof("\t\t %v", li)
			for _, po := range li.Line.Value.Points {
				log.Infof("\t\t\t %v", po)
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
