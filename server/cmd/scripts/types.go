package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/armsnyder/ts2go"
)

func main() {
	filePath := "/home/siddarth/dev/webdev/watch-party/shiba/client/app/lib/schema/chat.ts"
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	output := &bytes.Buffer{}

	err = ts2go.Generate(file, output)

	if err != nil {
		panic(err)
	}

	fmt.Print(output)

}
