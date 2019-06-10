/*
Copyright (C) 2019 Leo Tenenbaum

This file is part of AutoDistortion.

AutoDistortion is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

AutoDistortion is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with AutoDistortion.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
    "os"
    "fmt"
    "bufio"
    "strings"
    "flag"
    "time"
    "errors"
    "image"
    "image/png" // Image file types
    "image/jpeg"
    "math/rand"
    "github.com/pommicket/autodistortion/autodistortion"
)


func distort(filenameIn string, filenameOut string, functionLen int, nThreads int) error {
    reader, err := os.Open(filenameIn)
    if err != nil { return err }

    var outputFormat string
    if strings.HasSuffix(filenameOut, ".jpg") || strings.HasSuffix(filenameOut, ".jpeg") {
        outputFormat = "jpeg"
    } else if strings.HasSuffix(filenameOut, ".png") {
        outputFormat = "png"
    } else {
        return errors.New(fmt.Sprintf("Did not recognize format of: %v", filenameOut))

    }

    // Create it early so that if there's an error, we know sooner rather than
    // later
    outFile, err := os.Create(filenameOut)
    if err != nil { return err }
    defer outFile.Close()

    image, _, err := image.Decode(reader)
    reader.Close()
    if err != nil { return err }

    outImage := autodistortion.Distort(image, functionLen, nThreads)
    if err != nil { return err }

    switch outputFormat {
    case "png":
        err = png.Encode(outFile, outImage)
    case "jpeg":
        err = jpeg.Encode(outFile, outImage, nil)
    }
    return err
}

func main() {
    seed := flag.Int64("seed", time.Now().UTC().UnixNano(), "The seed to use for the random number generator")
    var filenameIn, filenameOut string
    flag.StringVar(&filenameIn, "in", "user input", "Which file should be distorted")
    flag.StringVar(&filenameOut, "out", "[in]_distorted.[extension]", "Which file to output to")
    fLength := flag.Int("function-len", 40, "The length of the distortion functions")
    threads := flag.Int("threads", 64, "Number of threads to use")
    quiet := flag.Bool("quiet", false, "Output nothing")
    flag.Parse()

    rand.Seed(*seed)
    if !*quiet {
        fmt.Println("Using seed:", *seed)
    }

    reader := bufio.NewReader(os.Stdin)
    if filenameIn == "user input" {

        if !*quiet {
            fmt.Print("What file would you like to distort? ")
        }

        text, err := reader.ReadString('\n')
        if err != nil {
            fmt.Println(err)
            return
        }
        filenameIn = strings.TrimSpace(text) // Remove newline
    }

    if filenameOut == "[in]_distorted.[extension]" {
        // Set filenameOut to foo.png -> foo_distorted.png
        dotPos := strings.IndexByte(filenameIn, '.')
        if dotPos == -1 {
            fmt.Println("Error: That file doesn't have a file extension!")
            return
        }
        filenameOut = filenameIn[:dotPos] + "_distorted" + filenameIn[dotPos:]
    }
    if !*quiet { fmt.Println("Loading...") }
    err := distort(filenameIn, filenameOut, *fLength, *threads)
    if err != nil {
        if !*quiet { fmt.Println("Error:", err) }
        return
    }
    if !*quiet { fmt.Println("Done! Output:", filenameOut) }
}