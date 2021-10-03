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

	getter := NewEmailGetter(20)

	getter.DebugMode = debugMode
	getter.PageNumber = pageNumber

	if onlyUsernames {
		getter.OnlyUsers = true
	}

	getter.RetrieveEmail(username)

	if following {
		getter.RetrieveFollowing(username)
	} else if followers {
		getter.RetrieveFollowers(username)
	}

	getter.wg.Wait()
}
