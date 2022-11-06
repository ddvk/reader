package main

import (
	"fmt"
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
	log.Info("parsed: ", scene.Author)
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

type A struct {
	foo int
}
type B struct {
	*A
	Bar int
}

func doStuff(s any) {
	s, ok := s.(*A)
	if ok {
		fmt.Println("it is a")

	}
	_, ok = s.(*B)
	if ok {
		fmt.Println("it is b")

	}

}

type C struct {
}

func Test[T *any]() T {
	return nil
}
func main() {
	prefixed := &prefixed.TextFormatter{
		DisableColors:   false,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceFormatting: true,
	}
	log.SetFormatter(prefixed)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
	err := _main()
	if err != nil {
		log.Fatal(err)
	}
}
