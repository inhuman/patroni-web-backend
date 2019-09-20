package utils

import (
	"fmt"
	"github.com/hokaccha/go-prettyjson"
	"github.com/sparrc/go-ping"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func PrettyPrintStruct(strct interface{}) {

	s, _ := prettyjson.Marshal(strct)
	log.Println(string(s))
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func CreateTempFile(data []byte, ext string) (*os.File, error) {

	path := filepath.Join(os.TempDir(), fmt.Sprintf("file.%d%s", time.Now().UnixNano(), ext))
	fh, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	_, err = fh.Write(data)
	if err != nil {
		return nil, err
	}

	return fh, nil
}

func GetPinger(uri string) (*ping.Pinger, error) {

	uri = "http:/" + uri

	u, err := url.Parse(uri)

	if err != nil {
		fmt.Printf("url parse error: %s\n", err)
	}

	fmt.Printf("PingHost url: %+v\n", u)

	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		fmt.Printf("split host error: %s\n", err)
	}

	pinger, err := ping.NewPinger(host)

	if err != nil {
		return nil, err
	}

	//pinger.SetPrivileged(true)
	return pinger, nil
}
