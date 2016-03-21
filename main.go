package main

import (
	"flag"
	"fmt"
	"sync"
)

var username = flag.String("username", "", "Username to execute the query")
var followers = flag.Bool("followers", false, "Get emails of follower users")
var following = flag.Bool("following", false, "Get emails of following users")
var noemails = flag.Bool("noemails", false, "Get the usernames instead of the emails")
var page = flag.Int("page", 1, "Page number for following and followers")

func main() {
	flag.Usage = func() {
		fmt.Println("E-Mail Getter")
		fmt.Println("  http://cixtor.com/")
		fmt.Println("  https://github.com/cixtor/emailgetter")
		fmt.Println("  https://en.wikipedia.org/wiki/Email_address_harvesting")
		fmt.Println("Usage:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *username == "" {
		fmt.Println("Missing username to query")
		flag.Usage()
	}

	var wg sync.WaitGroup
	var getter EmailGetter

	getter.PageNumber = *page

	if *noemails {
		getter.OnlyUsers = true
	}

	wg.Add(1) /* At least wait for one */
	go getter.RetrieveEmail(&wg, *username)

	if *following {
		getter.RetrieveFollowing(&wg, *username)
	} else if *followers {
		getter.RetrieveFollowers(&wg, *username)
	}

	wg.Wait()

	getter.PrintEmails()
}
