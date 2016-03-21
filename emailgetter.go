package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type EmailGetter struct {
	Addresses  []string
	RateLimit  bool
	OnlyUsers  bool
	PageNumber int
}

func (e *EmailGetter) RetrieveEmail(wg *sync.WaitGroup, username string) {
	defer wg.Done()

	if e.OnlyUsers {
		fmt.Println(username)
		return
	}

	/* Try to get it from the API */
	found := e.ExtractFromAPI(username)

	if found == false {
		/* Try to get it from the profile page */
		found = e.ExtractFromProfile(username)

		if found == false {
			/* Try to get it from the events endpoint */
			found = e.ExtractFromActivity(username)
		}
	}
}

func (e *EmailGetter) RetrieveFollowers(wg *sync.WaitGroup, username string) {
	e.FriendEmails(wg, username, "followers")
}

func (e *EmailGetter) RetrieveFollowing(wg *sync.WaitGroup, username string) {
	e.FriendEmails(wg, username, "following")
}

func (e *EmailGetter) FriendEmails(wg *sync.WaitGroup, username string, group string) {
	if e.PageNumber > 1 {
		group += "?page=" + strconv.Itoa(e.PageNumber)
	}

	content := e.Request("https://github.com/" + username + "/" + group)
	pattern := regexp.MustCompile(`<img alt="@([^"]+)"`)
	friends := pattern.FindAllStringSubmatch(string(content), -1)

	for _, data := range friends {
		if data[1] != username {
			wg.Add(1) /* Add more emails */
			go e.RetrieveEmail(wg, data[1])
		}
	}
}

func (e *EmailGetter) ExtractFromAPI(username string) bool {
	/* Skip if API is rate limited */
	if e.RateLimit == true {
		return false
	}

	content := e.Request("https://api.github.com/users/" + username)
	output := string(content) /* Convert to facilitate readability */

	if strings.Contains(output, "rate limit exceeded") {
		e.RateLimit = true
		return false
	}

	pattern := regexp.MustCompile(`"email": "([^"]+)",`)
	data := pattern.FindStringSubmatch(output)

	if len(data) == 2 && data[1] != "" {
		return e.AppendEmail(data[1])
	}

	return false
}

func (e *EmailGetter) ExtractFromProfile(username string) bool {
	content := e.Request("https://github.com/" + username)
	pattern := regexp.MustCompile(`"mailto:([^"]+)"`)
	data := pattern.FindStringSubmatch(string(content))

	if len(data) == 2 && data[1] != "" {
		var urlEncoded string = data[1]

		urlEncoded = strings.Replace(urlEncoded, ";", "", -1)
		urlEncoded = strings.Replace(urlEncoded, "&#x", "%", -1)

		if out, err := url.QueryUnescape(urlEncoded); err == nil {
			return e.AppendEmail(string(out))
		}
	}

	return false
}

func (e *EmailGetter) ExtractFromActivity(username string) bool {
	/* Skip if API is rate limited */
	if e.RateLimit == true {
		return false
	}

	content := e.Request("https://api.github.com/users/" + username + "/repos?type=owner&sort=updated")
	pattern := regexp.MustCompile(`"full_name": "([^"]+)",`)
	data := pattern.FindStringSubmatch(string(content))

	if len(data) == 2 && data[1] != "" {
		commits := e.Request("https://api.github.com/repos/" + data[1] + "/commits")
		expression := regexp.MustCompile(`"email": "([^"]+)",`)
		matches := expression.FindAllStringSubmatch(string(commits), -1)

		for _, match := range matches {
			e.AppendEmail(match[1])
		}

		return len(matches) > 0
	}

	return false
}

func (e *EmailGetter) Request(url string) []byte {
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

func (e *EmailGetter) AppendEmail(email string) bool {
	var isAlreadyAdded bool = false

	for _, item := range e.Addresses {
		if item == email {
			isAlreadyAdded = true
			break
		}
	}

	if isAlreadyAdded == false {
		e.Addresses = append(e.Addresses, email)
	}

	return true
}

func (e *EmailGetter) PrintEmails() {
	for _, email := range e.Addresses {
		fmt.Println(email)
	}
}
