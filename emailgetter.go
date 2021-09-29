package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// EmailGetter defines the structure that will contain the email addresses found
// on every visited page, whether the API service is limiting the HTTP requests
// or not, whether the program is limited to scan the users and the page number
// in case the user decided to scan the followers or following pages.
type EmailGetter struct {
	Addresses  []string
	RateLimit  bool
	OnlyUsers  bool
	PageNumber int
}

// RetrieveEmail contacts multiple websites looking for a valid email address
// that may be associated to the submitted username. At first, the program will
// try to find a valid email address in the (rate-limited) public API service,
// then will scan the user's profile page which sometimes exposes an email
// approved by the user itself encoded in hexadecimal format, finally, it will
// scan the other public API service (if the rate limit is not exceeded) for the
// most recent activity in any of the repositories managed by that account, then
// will extract all the valid email addresses from that output.
func (e *EmailGetter) RetrieveEmail(wg *sync.WaitGroup, username string) {
	defer wg.Done()

	if e.OnlyUsers {
		fmt.Println(username)
		return
	}

	if e.ExtractFromAPI(username) {
		// Try to get it from the API.
		return
	}

	if e.ExtractFromProfile(username) {
		// Try to get it from the profile page.
		return
	}

	e.ExtractFromActivity(username)
}

// RetrieveFollowers will try to find a valid email address for all the user
// accounts that are following the submitted username. By default, each full
// page in the followers section has 50 entries at a maximum.
func (e *EmailGetter) RetrieveFollowers(wg *sync.WaitGroup, username string) {
	e.FriendEmails(wg, username, "followers")
}

// RetrieveFollowing will try to find a valid email address for all the user
// accounts that are being followed by the submitted username. By default, each
// full page in the followers section has 50 entries at a maximum.
func (e *EmailGetter) RetrieveFollowing(wg *sync.WaitGroup, username string) {
	e.FriendEmails(wg, username, "following")
}

// FriendEmails scrappes the content of a public user's profile in search for a
// valid hexadecimal encoded email address. This operation has lower reliability
// than the other methods because the source of data comes from a setting in the
// user's account that allows him to either hide the email or submit any valid
// address to be seen by the public, so in many cases the information provided
// here is not accurate.
func (e *EmailGetter) FriendEmails(wg *sync.WaitGroup, username string, group string) {
	if e.PageNumber > 1 {
		group += "?page=" + strconv.Itoa(e.PageNumber)
	}

	content, err := e.Request("https://github.com/" + username + "/" + group)

	if err != nil {
		return
	}

	pattern := regexp.MustCompile(`<img alt="@([^"]+)"`)
	friends := pattern.FindAllStringSubmatch(string(content), -1)

	for _, data := range friends {
		if data[1] != username {
			wg.Add(1) /* Add more emails */
			go e.RetrieveEmail(wg, data[1])
		}
	}
}

// ExtractFromAPI sends a HTTP request to the public API service and, if not
// rate-limited, scrappes any valid email address from the JSON-encoded
// response. Notice that any valid email address found in this page will be
// returned as the result of this operation due to the impossibility of the
// program to determine which address is the real user's email.
func (e *EmailGetter) ExtractFromAPI(username string) bool {
	/* Skip if API is rate limited */
	if e.RateLimit {
		return false
	}

	content, err := e.Request("https://api.github.com/users/" + username)

	if err != nil {
		return false
	}

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

// ExtractFromProfile will, in case that the public API services are rate-
// limited, scan the user's profile page, find an hexadecimal encoded email
// address and decode it to a human readable string.
func (e *EmailGetter) ExtractFromProfile(username string) bool {
	content, err := e.Request("https://github.com/" + username)

	if err != nil {
		return false
	}

	pattern := regexp.MustCompile(`"mailto:([^"]+)"`)
	data := pattern.FindStringSubmatch(string(content))

	if len(data) == 2 && data[1] != "" {
		urlEncoded := data[1]

		urlEncoded = strings.Replace(urlEncoded, ";", "", -1)
		urlEncoded = strings.Replace(urlEncoded, "&#x", "%", -1)

		if out, err := url.QueryUnescape(urlEncoded); err == nil {
			return e.AppendEmail(out)
		}
	}

	return false
}

// ExtractFromActivity will read and extract every valid email address from the
// user's activity endpoint in the public API. This endpoint contains
// information about recent commits, recent pull-requests, recent issues created
// and commented by the user as well as additional information that, in certain
// cases, might be considered sensitive.
func (e *EmailGetter) ExtractFromActivity(username string) bool {
	/* Skip if API is rate limited */
	if e.RateLimit {
		return false
	}

	content, err := e.Request("https://api.github.com/users/" + username + "/repos?type=owner&sort=updated")

	if err != nil {
		return false
	}

	pattern := regexp.MustCompile(`"full_name": "([^"]+)",`)
	data := pattern.FindStringSubmatch(string(content))

	if len(data) == 2 && data[1] != "" {
		commits, err := e.Request("https://api.github.com/repos/" + data[1] + "/commits")

		if err != nil {
			return false
		}

		expression := regexp.MustCompile(`"email": "([^"]+)",`)
		matches := expression.FindAllStringSubmatch(string(commits), -1)

		for _, match := range matches {
			e.AppendEmail(match[1])
		}

		return len(matches) > 0
	}

	return false
}

var httpClient = http.Client{Timeout: time.Minute}

// Request sends a HTTP GET request to the URL passed in the parameters.
func (e *EmailGetter) Request(target string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, target, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (KHTML, like Gecko) Safari/537.36")

	res, err := httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	reader := io.LimitReader(res.Body, 2<<20)

	return ioutil.ReadAll(reader)
}

// AppendEmail will insert a new entry into the email address list.
func (e *EmailGetter) AppendEmail(email string) bool {
	var isAlreadyAdded bool

	for _, item := range e.Addresses {
		if item == email {
			isAlreadyAdded = true
			break
		}
	}

	if !isAlreadyAdded {
		e.Addresses = append(e.Addresses, email)
	}

	return true
}

// PrintEmails will send all the collected emails to os.Stdout
func (e *EmailGetter) PrintEmails() {
	for _, email := range e.Addresses {
		fmt.Println(email)
	}
}
