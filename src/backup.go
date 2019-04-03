package main

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

//Panic if an error
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func zipFiles(filename string, files []string) error {

	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	//Add files to zip
	for _, file := range files {
		if err = addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {

	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	//Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	//Using FileInfoHeader() above only uses the basename of the file. If we want
	//to preserve the folder structure we can overwrite this with the full path.
	header.Name = filename

	//Change to deflate to gain better compression
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

func walk(basePath string, walkedFiles *[]string) error {
	//Open the Directory
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		if file.IsDir() {
			//Recurse
			newBase := filepath.Join(basePath, file.Name()) + "/"
			err = walk(newBase, walkedFiles)
			if err != nil {
				return err
			}
		} else {
			*walkedFiles = append(*walkedFiles, filepath.Join(basePath, file.Name()))
		}
	}
	return nil
}

/*
Backs up a Minecraft world. Takes a path to the
server folder as an argument.
*/
func Backup(path, world string) error {
	os.Chdir(path)
	files, err := ioutil.ReadDir(".")
	if err != nil {
		return err
	}

	worldFound := false
	for _, file := range files {
		if file.Name() == world {
			worldFound = true
			fi, err := os.Stat(file.Name())
			if err != nil {
				return err
			}
			if fi.Mode().IsDir() {
				filesToZip := []string{}
				err = walk(file.Name(), &filesToZip)

				fmt.Println("Zipping")
				err = zipFiles(world+".zip", filesToZip)
				if err != nil {
					return err
				}
			}
		}
	}
	if !worldFound {
		err = errors.New("World file not found")
	}
	return err
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	scanner := bufio.NewScanner(reader)
	scanner.Scan()
	text := scanner.Text()
	err := Backup(text, "world")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Zipped!")
	}
}
