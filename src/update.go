package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

//Panic if an error
func check(e error) {
	if e != nil {
		panic(e)
	}
}

//See if a file exists
func checkFile(fileName string) (bool, error) {
	//Never tested the pathSlice code properly

	//Split up the path of the file and reassemble it
	dir := ""
	pathSlice := strings.Split(fileName, "/")
	if pathSlice[0] == fileName {
		dir = "."
	} else {
		dir = "/"
		for i := 0; i < len(pathSlice); i++ {
			dir += pathSlice[i] + "/"
		}
	}

	//Find files in the directory
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return false, err
	}

	//Compare their names to our given name
	for i := 0; i < len(files); i++ {
		if files[i].Name() == fileName {
			return true, nil
		}
	}
	return false, nil
}

//Compare two files byte by byte
func fileCmp(fileName1 string, fileName2 string, num int) (bool, error) {
	//Open the file
	file1, err := os.Open(fileName1)
	defer file1.Close()
	if err != nil {
		return false, err
	}

	//Create a new byte reader
	reader1 := bufio.NewReader(file1)
	bytes1 := make([]byte, num)

	//Open the file
	file2, err := os.Open(fileName2)
	defer file2.Close()
	if err != nil {
		return false, err
	}

	//Create a new byte reader
	reader2 := bufio.NewReader(file2)
	bytes2 := make([]byte, num)

	//Loop while not the end of the file and they files are the same
	end := false
	for !end {
		//Read in first file's bytes
		num1, err := reader1.Read(bytes1)
		if err != nil && err != io.EOF {
			return false, err
		} else if err == io.EOF {
			end = true
		}

		//Read in second file's bytes
		num2, err := reader2.Read(bytes2)
		if err != nil && err != io.EOF {
			return false, err
		} else if err == io.EOF {
			end = true
		}

		//Check if byte count is the same
		if num1 != num2 {
			return false, err
		}

		if bytes.Compare(bytes1, bytes2) != 0 {
			return false, err
		}
	}
	return true, err
}

func getURL(url string) (string, error) {
	//Set the transport type and create the client
	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
	client := &http.Client{Transport: tr}

	//Do an https get
	res, err := client.Get(url)
	defer res.Body.Close()
	if err != nil {
		return "", err
	}

	//Read all of the data
	byte_str, err := ioutil.ReadAll(res.Body)
	text := string(byte_str)
	if err != nil {
		return text, err
	} else {
		return text, nil
	}
}

func downloadFile(filepath, url string) error {
	// Get the data
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	// Create the file
	out, err := os.Create(filepath)
	defer out.Close()
	if err != nil {
		return err
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

/*
This function updates a vanilla 'server.jar' file in a directory.
It first goes to the Minecraft website and then uses regular
expressions to grab a link to the server download page.
It then downloads the file, and if a 'server.jar' doesn't exist,
it downloads the new file as 'new_server.jar', and then renames
it to 'server.jar'. If 'server.jar' does exist, it downloads the new
one, then compares it to the old one. If they're the same, it throws
out the old one. If they're not, it timestamps the old one and saves
it with that timestamp, then renames the new server to 'server.jar'
*/
func UpdateServer() {
	server := "server.jar"
	newServer := "new_server.jar"
	url := "https://www.minecraft.net/en-us/download/server/"

	//getURL comes from web.go
	text, err := getURL(url)
	if err != nil {
		fmt.Println("Couldn't find find", url)
		return
	}

	//Look for the server jar
	re, err := regexp.Compile("(http|https):\\/\\/.*server\\.jar")
	if err != nil {
		fmt.Println("Bad regex")
		return
	}

	//Extract the url
	url = re.FindString(text)

	//downloadFile comes from web.go
	err = downloadFile(newServer, url)
	check(err)

	//checkFile comes from fileio.go
	readByte, err := checkFile(server)
	check(err)

	if readByte {
		//fileCmp comes from fileio.go
		//Compare the two server files at ever 256 bytes
		readByte, err = fileCmp(server, newServer, 256)
		check(err)
		if !readByte {
			//Produces a formatted time string
			timeString := time.Now().String()

			//Grab the date
			re, err = regexp.Compile("[0-9]{4}(-[0-9]{2}){2}")
			date := re.FindString(timeString)

			//Grab the time
			re, err = regexp.Compile("([0-9]{2}:){2}[0-9]{2}")
			currTime := re.FindString(timeString)

			timeStamp := date + "_" + currTime
			os.Rename(server, timeStamp+"_"+server)
		}
	}
	os.Rename(newServer, server)
}

func main() {
	UpdateServer()
}
