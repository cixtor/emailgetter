package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
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
	wg         sync.WaitGroup
	reqLock    chan bool
	Addresses  sync.Map
	OnlyUsers  bool
	DebugMode  bool
	PageNumber int
}

var reEmail = regexp.MustCompile(`"email": "([^"]+)",`)
var reImage = regexp.MustCompile(`<img alt="@([^"]+)"`)
var reMailto = regexp.MustCompile(`"mailto:([^"]+)"`)
var reFullname = regexp.MustCompile(`"full_name": "([^"]+)",`)

func NewEmailGetter(maxRequests int) *EmailGetter {
	return &EmailGetter{
		reqLock:   make(chan bool, maxRequests),
		Addresses: sync.Map{},
	}
}

// RetrieveEmail contacts multiple websites looking for a valid email address
// that may be associated to the submitted username. At first, the program will
// try to find a valid email address in the (rate-limited) public API service,
// then will scan the user's profile page which sometimes exposes an email
// approved by the user itself encoded in hexadecimal format, finally, it will
// scan the other public API service (if the rate limit is not exceeded) for the
// most recent activity in any of the repositories managed by that account, then
// will extract all the valid email addresses from that output.
func (e *EmailGetter) RetrieveEmail(username string) {
	e.wg.Add(1)
	e.reqLock <- true
	defer e.wg.Done()
	defer func() { <-e.reqLock }()

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
func (e *EmailGetter) RetrieveFollowers(username string) {
	e.FriendEmails(username, "followers")
}

// RetrieveFollowing will try to find a valid email address for all the user
// accounts that are being followed by the submitted username. By default, each
// full page in the followers section has 50 entries at a maximum.
func (e *EmailGetter) RetrieveFollowing(username string) {
	e.FriendEmails(username, "following")
}

// FriendEmails scrappes the content of a public user's profile in search for a
// valid hexadecimal encoded email address. This operation has lower reliability
// than the other methods because the source of data comes from a setting in the
// user's account that allows him to either hide the email or submit any valid
// address to be seen by the public, so in many cases the information provided
// here is not accurate.
func (e *EmailGetter) FriendEmails(username string, group string) {
	if e.PageNumber > 1 {
		group += "?page=" + strconv.Itoa(e.PageNumber)
	}

	content, err := e.Request("https://github.com/" + username + "/" + group)

	if err != nil {
		return
	}

	friends := reImage.FindAllStringSubmatch(string(content), -1)

	for _, data := range friends {
		if data[1] == username {
			continue
		}

		go e.RetrieveEmail(data[1])
	}
}

// ExtractFromAPI sends a HTTP request to the public API service and, if not
// rate-limited, scrappes any valid email address from the JSON-encoded
// response. Notice that any valid email address found in this page will be
// returned as the result of this operation due to the impossibility of the
// program to determine which address is the real user's email.
func (e *EmailGetter) ExtractFromAPI(username string) bool {
	out, err := e.Request("https://api.github.com/users/" + username)

	if err != nil {
		return false
	}

	data := reEmail.FindSubmatch(out)

	// Minimal email address is x@y
	if len(data) == 2 && len(data[1]) >= 3 {
		return e.PrintEmail(string(data[1]))
	}

	return false
}

// ExtractFromProfile will, in case that the public API services are rate-
// limited, scan the user's profile page, find an hexadecimal encoded email
// address and decode it to a human readable string.
func (e *EmailGetter) ExtractFromProfile(username string) bool {
	out, err := e.Request("https://github.com/" + username)

	if err != nil {
		return false
	}

	data := reMailto.FindSubmatch(out)

	if len(data) < 2 || len(data[1]) < 3 {
		return false
	}

	s := string(data[1])
	s = strings.Replace(s, ";", "", -1)
	s = strings.Replace(s, "&#x", "%", -1)

	clean, err := url.QueryUnescape(s)

	if err != nil {
		return false
	}

	return e.PrintEmail(clean)
}

// ExtractFromActivity will read and extract every valid email address from the
// user's activity endpoint in the public API. This endpoint contains
// information about recent commits, recent pull-requests, recent issues created
// and commented by the user as well as additional information that, in certain
// cases, might be considered sensitive.
func (e *EmailGetter) ExtractFromActivity(username string) bool {
	out, err := e.Request("https://api.github.com/users/" + username + "/repos?type=owner&sort=updated")

	if err != nil {
		return false
	}

	matches := reFullname.FindAllSubmatch(out, -1)

	for _, match := range matches {
		if len(match) != 2 {
			continue
		}

		e.ExtractFromCommits(string(match[1]))
	}

	return false
}

func (e *EmailGetter) ExtractFromCommits(repo string) {
	commits, err := e.Request("https://api.github.com/repos/" + repo + "/commits")

	if err != nil {
		return
	}

	matches := reEmail.FindAllSubmatch(commits, -1)

	for _, match := range matches {
		if len(match) != 2 {
			continue
		}

		e.PrintEmail(string(match[1]))
	}
}

var httpClient = http.Client{Timeout: time.Minute}

// Request sends a HTTP GET request to the URL passed in the parameters.
func (e *EmailGetter) Request(target string) ([]byte, error) {
	if e.DebugMode {
		fmt.Println(target)
	}

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
	out, err := ioutil.ReadAll(reader)

	if err != nil {
		return nil, err
	}

	// Exceeding the rate limit:
	//
	// > HTTP/2 403
	// > Date: Tue, 20 Aug 2013 14:50:41 GMT
	// > x-ratelimit-limit: 60
	// > x-ratelimit-remaining: 0
	// > x-ratelimit-used: 60
	// > x-ratelimit-reset: 1377013266
	// >
	// > {"message":"API rate limit exceeded for 0.0.0.0.", ...}
	//
	// Source: https://docs.github.com/rest/overview/resources-in-the-rest-api#rate-limiting
	if res.StatusCode == http.StatusForbidden {
		ts, _ := strconv.ParseInt(res.Header.Get("X-RateLimit-Reset"), 10, 64)
		fmt.Printf("rate limit exceeded; try again at %s\n", time.Unix(ts, 0))
		os.Exit(1)
		return nil, nil
	}

	if bytes.Contains(out, []byte("rate limit exceeded")) {
		fmt.Printf("%s\n", out)
		os.Exit(1)
		return nil, nil
	}

	return out, nil
}

// PrintEmail writes an email address to /dev/stdout if unique.
func (e *EmailGetter) PrintEmail(email string) bool {
	if strings.HasSuffix(email, "@users.noreply.github.com") {
		return false
	}

	// Skip if the email has already been printed before.
	if _, ok := e.Addresses.Load(email); ok {
		return false
	}

	e.Addresses.Store(email, true)

	fmt.Println(email)

	return true
}
