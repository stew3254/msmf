package main

import (
	"fmt"
	"os"
	"regexp"
	"time"
)

//Panic if an error
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	server := "server.jar"
	newServer := "new_server.jar"
	url := "https://www.minecraft.net/en-us/download/server/"
	text, err := GetURL(url)
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

	err = DownloadFile(newServer, url)
	check(err)

	b, err := CheckFile(server)
	check(err)

	if b {
		//Compare the two server files at ever 256 bytes
		b, err = FileCmp(server, newServer, 256)
		check(err)
		if !b {
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
