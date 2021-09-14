package main

import (
	"encoding/base64"
	"io/ioutil"
)

func main() {
	data, err := ioutil.ReadFile("./工商银行.xjs")
	if err != nil {
		panic(err)
	}

	decoded, err := base64.StdEncoding.DecodeString(string(data))
	// decodestr := string(decoded)
	// fmt.Println(decodestr, err)

	ioutil.WriteFile("./工商银行.jar", decoded, 0755)
}
