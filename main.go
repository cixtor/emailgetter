package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
)

var debugMode bool
var followers bool
var following bool
var onlyUsernames bool
var pageNumber int

func main() {
	flag.BoolVar(&debugMode, "debug", false, "Print the URLs that are being requested")
	flag.BoolVar(&followers, "followers", false, "Print the email(s) associated to each follower")
	flag.BoolVar(&following, "following", false, "Print the email(s) associated to each following")
	flag.BoolVar(&onlyUsernames, "only-usernames", false, "Print following/followers usernames instead of emails")
	flag.IntVar(&pageNumber, "page", 1, "Page number for following and followers")

	flag.Usage = func() {
		fmt.Println("usage: emailgetter [OPTIONS] [USERNAME]")
		flag.PrintDefaults()
	}

	flag.Parse()

	username := flag.Arg(0)

	if username == "" {
		flag.Usage()
		os.Exit(2)
	}

	var wg sync.WaitGroup
	var getter EmailGetter

	getter.DebugMode = debugMode
	getter.PageNumber = pageNumber

	if onlyUsernames {
		getter.OnlyUsers = true
	}

	wg.Add(1) /* At least wait for one */
	go getter.RetrieveEmail(&wg, username)

	if following {
		getter.RetrieveFollowing(&wg, username)
	} else if followers {
		getter.RetrieveFollowers(&wg, username)
	}

	wg.Wait()

	getter.PrintEmails()
}
