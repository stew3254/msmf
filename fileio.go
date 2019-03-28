package main

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

//See if a file exists
func CheckFile(fileName string) (bool, error) {
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
func FileCmp(fileName1 string, fileName2 string, num int) (bool, error) {
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
