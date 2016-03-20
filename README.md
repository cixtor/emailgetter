### E-Mail Getter

Command line tool to extract email addresses associated to a specific username from certain websites. This project was inspired by [github-email](https://github.com/paulirish/github-email) a shell script that communicates with GitHub's API service to retrieve the email address associated to an account. This project aims to offer the same functionality and expand to other popular websites to add more reliability; notice that in comparison with the original idea this version of the code will delete duplicate entries as well as clear non-responsive addresses.

### Installation

I do not distribute binaries for security reasons.

```shell
go get -u github.com/cixtor/emailgetter
```

### Usage

```
$ emailgetter -help
Usage:
  -username  string Username to execute the query
  -noemails  bool   Get the usernames instead of the emails
  -following bool   Get emails of users that [username] is following
  -followers bool   Get emails of users that are following [username]
  -page      int    Page number for following and followers queries
$ emailgetter -username cixtor
$ emailgetter -username cixtor -following
$ emailgetter -username cixtor -followers -page 5
$ emailgetter -username cixtor -followers -page 5 -noemails
```

> Email harvesting is the process of obtaining lists of email addresses using various methods for use in bulk email. People may harvest email addresses from a number of sources. A popular method uses email addresses which their owners have published for other purposes. Simply searching the web for pages with addresses — such as corporate staff directories or membership lists of professional societies — using automated programs can yield thousands of addresses, most of them deliverable. The DNS and WHOIS systems require the publication of technical contact information for all Internet domains; people have illegally trawled these resources for email addresses.
>
> — https://en.wikipedia.org/wiki/Email_address_harvesting  
> — https://en.wikipedia.org/wiki/Anti-spam_techniques
