package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
)

var username = flag.String("u", "", "Username to query")

func httpGetRequest(url string) []byte {
	client := http.Client{}

	req, err := http.NewRequest("GET", url, nil)

	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Accept-Language", "en-US,en;q=0.8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (KHTML, like Gecko) Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	if err != nil {
		panic(err)
	}

	resp, err := client.Do(req)

	defer resp.Body.Close()

	// I understand that ioutil.ReadAll is bad.
	content, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	return content
}

func extractFromAPI(username string) (string, bool) {
	content := httpGetRequest("https://api.github.com/users/" + username)
	pattern := regexp.MustCompile(`"email": "([^"]+)",`)
	data := pattern.FindStringSubmatch(string(content))

	if len(data) == 2 && data[1] != "" {
		return data[1], true
	}

	return "", false
}

func printProfileEmail(username string) {
	var email string = ""
	var found bool = false

	email, found = extractFromAPI(username)

	if found == true {
		fmt.Println(email)
		return
	}
}

func main() {
	flag.Parse()

	flag.Usage = func() {
		fmt.Println("E-Mail Getter")
		fmt.Println("http://cixtor.com/")
		fmt.Println("https://github.com/cixtor/emailgetter")

		flag.PrintDefaults()

		os.Exit(2)
	}

	if *username == "" {
		fmt.Println("Missing username to query")
		flag.Usage()
	}

	printProfileEmail(*username)
}
