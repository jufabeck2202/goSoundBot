package main

import (
	"fmt"
	"github.com/jonas747/dca"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func encode() {

		files, err := ioutil.ReadDir("./mp3")
		if err != nil {
			log.Fatal(err)
		}

		for i, f := range files {
			// Encoding a file and saving it to disk
			options := dca.StdEncodeOptions
			options.RawOutput = true
			encodeSession, err1 := dca.EncodeFile("./mp3/"+f.Name(), options)
			// Make sure everything is cleaned up, that for example the encoding process if any issues happened isnt lingering around
			defer encodeSession.Cleanup()
			if err1 != nil {
				fmt.Println(err1)
				// Handle the error
			}

			output, err := os.Create("./dca/"+strconv.Itoa(i)+".dca")
			if err != nil {
				fmt.Println(err)
				// Handle the error
			}

			io.Copy(output, encodeSession)
		}
	loadSounds()
}


