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

type EmailGetter struct {
	Addresses []string
	RateLimit bool
}

func (getter *EmailGetter) RetrieveEmail(username string) {
	var found bool = false

	if found = getter.ExtractFromAPI(username); found {
		return
	}

	if found = getter.ExtractFromProfile(username); found {
		return
	}

	if found = getter.ExtractFromActivity(username); found {
		return
	}
}

func (getter *EmailGetter) FriendEmails(username string) {
	content := getter.Request("https://github.com/" + username + "/following")
	pattern := regexp.MustCompile(`<img alt="@([^"]+)"`)
	friends := pattern.FindAllStringSubmatch(string(content), -1)

	for _, data := range friends {
		if data[1] != username {
			getter.RetrieveEmail(data[1])
		}
	}
}

func (getter *EmailGetter) ExtractFromAPI(username string) bool {
	/* Skip if API is rate limited */
	if getter.RateLimit == true {
		return false
	}

	content := getter.Request("https://api.github.com/users/" + username)
	output := string(content) /* Convert to facilitate readability */

	if strings.Contains(output, "rate limit exceeded") {
		getter.RateLimit = true
		return false
	}

	pattern := regexp.MustCompile(`"email": "([^"]+)",`)
	data := pattern.FindStringSubmatch(output)

	if len(data) == 2 && data[1] != "" {
		return getter.AppendEmail(data[1])
	}

	return false
}

func (getter *EmailGetter) ExtractFromProfile(username string) bool {
	content := getter.Request("https://github.com/" + username)
	pattern := regexp.MustCompile(`"mailto:([^"]+)"`)
	data := pattern.FindStringSubmatch(string(content))

	if len(data) == 2 && data[1] != "" {
		var urlEncoded string = data[1]

		urlEncoded = strings.Replace(urlEncoded, ";", "", -1)
		urlEncoded = strings.Replace(urlEncoded, "&#x", "%", -1)

		if out, err := url.QueryUnescape(urlEncoded); err == nil {
			return getter.AppendEmail(string(out))
		}
	}

	return false
}

func (getter *EmailGetter) ExtractFromActivity(username string) bool {
	/* Skip if API is rate limited */
	if getter.RateLimit == true {
		return false
	}

	content := getter.Request("https://api.github.com/users/" + username + "/repos?type=owner&sort=updated")
	pattern := regexp.MustCompile(`"full_name": "([^"]+)",`)
	data := pattern.FindStringSubmatch(string(content))

	if len(data) == 2 && data[1] != "" {
		commits := getter.Request("https://api.github.com/repos/" + data[1] + "/commits")
		expression := regexp.MustCompile(`"email": "([^"]+)",`)
		matches := expression.FindAllStringSubmatch(string(commits), -1)

		for _, match := range matches {
			getter.AppendEmail(match[1])
		}

		return len(matches) > 0
	}

	return false
}

func (getter *EmailGetter) Request(url string) []byte {
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

func (getter *EmailGetter) AppendEmail(email string) bool {
	var isAlreadyAdded bool = false

	for _, item := range getter.Addresses {
		if item == email {
			isAlreadyAdded = true
			break
		}
	}

	if isAlreadyAdded == false {
		getter.Addresses = append(getter.Addresses, email)
	}

	return true
}

func (getter *EmailGetter) PrintEmails() {
	for _, email := range getter.Addresses {
		fmt.Println(email)
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

	var getter EmailGetter

	getter.RetrieveEmail(*username)

	if *friends == true {
		getter.FriendEmails(*username)
	}

	getter.PrintEmails()
}
