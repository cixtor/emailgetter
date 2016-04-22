package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var username = flag.String("username", "", "Username to execute the query")
var friends = flag.Bool("friends", false, "Get emails of following users")

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

func extractFromProfile(username string) (string, bool) {
	content := httpGetRequest("https://github.com/" + username)
	pattern := regexp.MustCompile(`"mailto:([^"]+)"`)
	data := pattern.FindStringSubmatch(string(content))

	if len(data) == 2 && data[1] != "" {
		var urlEncoded string = data[1]

		urlEncoded = strings.Replace(urlEncoded, ";", "", -1)
		urlEncoded = strings.Replace(urlEncoded, "&#x", "%", -1)

		output, err := url.QueryUnescape(urlEncoded)

		if err != nil {
			return "", false
		}

		return string(output), true
	}

	return "", false
}

func extractFromActivity(username string) (string, bool) {
	content := httpGetRequest("https://api.github.com/users/" + username + "/repos?type=owner&sort=updated")
	pattern := regexp.MustCompile(`"full_name": "([^"]+)",`)
	data := pattern.FindStringSubmatch(string(content))

	if len(data) == 2 && data[1] != "" {
		var emails []string

		commits := httpGetRequest("https://api.github.com/repos/" + data[1] + "/commits")
		expression := regexp.MustCompile(`"email": "([^"]+)",`)
		matches := expression.FindAllStringSubmatch(string(commits), -1)

		for _, match := range matches {
			emails = append(emails, match[1])
		}

		return strings.Join(emails, "\n"), true
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

	email, found = extractFromProfile(username)

	if found == true {
		fmt.Println(email)
		return
	}

	email, found = extractFromActivity(username)

	if found == true {
		fmt.Println(email)
		return
	}
}

func printFriendEmails(username string) {
	content := httpGetRequest("https://github.com/" + username + "/following")
	pattern := regexp.MustCompile(`<img alt="@([^"]+)"`)
	friends := pattern.FindAllStringSubmatch(string(content), -1)

	for _, data := range friends {
		if data[1] != username {
			printProfileEmail(data[1])
		}
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

	if *friends == true {
		printFriendEmails(*username)
	}
}
