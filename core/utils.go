package core

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
)

//func IntToHex(n int64) []byte {
//	return []byte(strconv.FormatInt(n, 16))
//}

func FormatB(hash []byte) []byte {
	last := bytes.Join(
		[][]byte{[]byte("b"), hash},
		[]byte{},
	)
	return last
}

func BlockFile(h int) string {
	return  fmt.Sprintf("./blocks_gen/%d.dat", h)
}

func ChainExists() bool {
	files, err := ioutil.ReadDir("./blocks_gen")
	if err != nil {
		os.Mkdir("./blocks_gen", 0777)
	}
	return len(files) > 0
}