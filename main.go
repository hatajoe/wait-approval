package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"
)

const (
	approvalComment string = `/approve`
	denyComment     string = `/deny`
)

var (
	pollingInterval time.Duration = 10 * time.Second
	launchedAt      time.Time

	githubRepository        string
	githubRepositoryOwner   string
	githubToken             string
	githubPullRequestNumber int
)

func init() {
	launchedAt = time.Now()
	githubRepository = strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")[1]
	if githubRepository == "" {
		log.Fatal("GITHUB_REPOSITORY env value must be specified")
	}
	githubRepositoryOwner = os.Getenv("GITHUB_REPOSITORY_OWNER")
	if githubRepositoryOwner == "" {
		log.Fatal("GITHUB_REPOSITORY_OWNER env value must be specified")
	}
	githubToken = os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatal("GITHUB_TOKEN env value must be specified")
	}

	var err error
	githubPullRequestNumber, err = strconv.Atoi(os.Getenv("INPUT_GITHUB-PULL-REQUEST-NUMBER"))
	if err != nil {
		log.Fatalf("github-pull-request-number input value must be specified: %s", err)
	}
}

func main() {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	ch := make(chan error)
	ticker := time.NewTicker(pollingInterval)
	go func() {
		defer func() {
			ticker.Stop()
			close(ch)
		}()

		greeting := fmt.Sprintf("Please `/approve` or `/deny` workflow.")
		if _, _, err := client.Issues.CreateComment(ctx, githubRepositoryOwner, githubRepository, githubPullRequestNumber, &github.IssueComment{
			Body: &greeting,
		}); err != nil {
			ch <- err
			return
		}

		for {
			for range ticker.C {
				comments, _, err := client.Issues.ListComments(ctx, githubRepositoryOwner, githubRepository, githubPullRequestNumber, &github.IssueListCommentsOptions{})
				if err != nil {
					ch <- err
					return
				}
				for _, comment := range comments {
					if comment.CreatedAt.Before(launchedAt) {
						continue
					}

					body := comment.GetBody()
					if matched, err := regexp.MatchString(fmt.Sprintf("^%s$", approvalComment), body); matched && err == nil {
						comment := fmt.Sprintf("workflow has approved by %s.", comment.GetUser().GetLogin())
						if _, _, err := client.Issues.CreateComment(ctx, githubRepositoryOwner, githubRepository, githubPullRequestNumber, &github.IssueComment{
							Body: &comment,
						}); err != nil {
							ch <- err
							return
						}
						log.Print(comment)
						return
					}
					if matched, err := regexp.MatchString(fmt.Sprintf("^%s$", denyComment), body); matched && err == nil {
						comment := fmt.Sprintf("workflow has denied by %s.", comment.GetUser().GetLogin())
						if _, _, err := client.Issues.CreateComment(ctx, githubRepositoryOwner, githubRepository, githubPullRequestNumber, &github.IssueComment{
							Body: &comment,
						}); err != nil {
							ch <- err
							return
						}
						ch <- fmt.Errorf(comment)
						return
					}
				}
			}
		}
	}()

	for err := range ch {
		if err != nil {
			log.Fatal(err)
		}
	}
	os.Exit(0)
}
