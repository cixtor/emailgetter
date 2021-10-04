package main

import (
	"flag"
	"fmt"
	"os"
)

var debugMode bool
var followers bool
var following bool
var onlyUsernames bool
var pageNumber int
var maxThreads int

func main() {
	flag.BoolVar(&debugMode, "debug", false, "Print the URLs that are being requested")
	flag.BoolVar(&followers, "followers", false, "Print the email(s) associated to each follower")
	flag.BoolVar(&following, "following", false, "Print the email(s) associated to each following")
	flag.BoolVar(&onlyUsernames, "only-usernames", false, "Print following/followers usernames instead of emails")
	flag.IntVar(&pageNumber, "p", 1, "Page number for following and followers")
	flag.IntVar(&maxThreads, "t", 10, "Maximum number of concurrent HTTP requests")

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

	getter := NewEmailGetter(maxThreads)

	getter.DebugMode = debugMode
	getter.PageNumber = pageNumber
	getter.OnlyUsers = onlyUsernames

	getter.RetrieveEmail(username)

	if following {
		getter.RetrieveFollowing(username)
	}

	if followers {
		getter.RetrieveFollowers(username)
	}

	getter.wg.Wait()
}
