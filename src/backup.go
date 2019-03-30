package main

import (
  "bufio"
	"fmt"
  "io/ioutil"
  "os"
)

//Panic if an error
func check(e error) {
	if e != nil {
		panic(e)
	}
}

/*
Backs up a Minecraft world. Takes a path to the
server folder as an argument.
*/
func Backup(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
    return err
	}

	for _, file := range files {
    fmt.Println(file.Name())
  }
  return nil
}

func main() {
  reader := bufio.NewReader(os.Stdin)
  text, _ := reader.ReadString('\n')
  Backup(text)
}
