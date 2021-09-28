# E-Mail Getter

Command line tool to extract email addresses associated to a specific GitHub account.

This project was inspired by [github-email](https://github.com/paulirish/github-email) a shell script that communicates with GitHub's API service to retrieve the email address associated to an account. This project aims to offer the same functionality and expand to other popular websites to add more reliability; notice that in comparison with the original idea this version of the code will delete duplicate entries as well as clear non-responsive addresses.

## Warning

Please be mindful and do not use this program to spam people.

> Email harvesting is the process of obtaining lists of email addresses using various methods for use in bulk email. People may harvest email addresses from a number of sources. A popular method uses email addresses which their owners have published for other purposes. Simply searching the web for pages with addresses — such as corporate staff directories or membership lists of professional societies — using automated programs can yield thousands of addresses, most of them deliverable. The DNS and WHOIS systems require the publication of technical contact information for all Internet domains; people have illegally trawled these resources for email addresses.
>
> — https://en.wikipedia.org/wiki/Email_address_harvesting  
> — https://en.wikipedia.org/wiki/Anti-spam_techniques

## Installation

I do not distribute binaries for security reasons.

```shell
go get -u github.com/cixtor/emailgetter
```

## Usage

Print one or more publicly available emails associated to GitHub user [torvalds](http://github.com/torvalds):

```bash
emailgetter torvalds
```

Print one or more emails for the GitHub users who follow the same person:

```bash
emailgetter -followers torvalds
```

Restrict the results to a specific page:

```bash
emailgetter -followers -page 5 torvalds
```
