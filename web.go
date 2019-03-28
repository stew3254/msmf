package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func GetURL(url string) (string, error) {
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

func DownloadFile(filepath, url string) error {
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
