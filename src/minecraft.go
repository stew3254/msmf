package main

import (
  "bufio"
  "fmt"
  "log"
  "io"
  "os"
  "os/exec"
)

func input(in *io.WriteCloser) {
  defer (*in).Close()
  reader := bufio.NewReader(os.Stdin)
  for {
    io.Copy(*in, reader)
  }
}

func main() {
  fmt.Print("Input the directory to your server: ")
  //Read in directory
  reader := bufio.NewReader(os.Stdin)
  scanner := bufio.NewScanner(reader)
  scanner.Split(bufio.ScanLines)
  scanner.Scan()
  m := scanner.Text()

  err := os.Chdir(m)
  if err != nil {
    log.Fatal(err)
    return
  }

  cmd := exec.Command("java", "-Xmx1024M", "-Xms1024M", "-jar", "server.jar", "nogui")
  //cmd := exec.Command("./test.sh")
  stdin, err := cmd.StdinPipe()
  if err != nil {
    log.Fatal(err)
  }

  go input(&stdin)
  if err != nil {
    log.Fatal(err)
  }

  stdout, err := cmd.StdoutPipe()
  if err != nil {
    log.Fatal(err)
  }

  if err := cmd.Start(); err != nil {
    log.Fatal(err)
  }

  scanner = bufio.NewScanner(stdout)
  scanner.Split(bufio.ScanLines)
  for scanner.Scan() {
    m := scanner.Text()
    fmt.Println(m)
  }
  //cmd.Wait()
}
