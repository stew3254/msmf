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

  //GetURL comes from web.go
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

  //DownloadFile comes from web.go
	err = DownloadFile(newServer, url)
	check(err)

  //CheckFile comes from fileio.go
	readByte, err := CheckFile(server)
	check(err)

	if readByte {
		//FileCmp comes from fileio.go
		//Compare the two server files at ever 256 bytes
		readByte, err = FileCmp(server, newServer, 256)
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
