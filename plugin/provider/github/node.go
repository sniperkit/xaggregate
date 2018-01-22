package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
)

var (
	ctx      context.Context
	origin   string
	target   string
	seen     = make(map[string]bool, 0) // concurrent map ?!
	done     = make(chan *UserNode)
	jobQueue = make(chan jobRequest, 1000)
)

// UserNode represents a single node in the follower graph. The Parent
// field holds the user that discovered this node.
type UserNode struct {
	Login  string
	Parent *UserNode
	Page   int
}

// newUserNode creates/returns a new *UserNode and adds this login to the seen
// map. If we've already seen this login, return nil.
func newUserNode(login string, parent *UserNode) *UserNode {
	if _, ok := seen[login]; ok {
		return nil
	}
	seen[login] = true

	return &UserNode{
		Login:  login,
		Parent: parent,
		Page:   1,
	}
}

func (node *UserNode) String() string {
	if node.Parent != nil {
		return fmt.Sprintf("%s -> %s", node.Parent.String(), node.Login)
	}
	return node.Login
}

// run retrieves the list of users that the current user is following and
// enqueues each of them.
func (node *UserNode) Process() {
	opts := github.ListOptions{Page: node.Page, PerPage: 100}
	following, _, err := ghClient.Users.ListFollowing(ctx, node.Login, &opts)
	if err != nil {
		log.Fatal(err)
	}

	for _, followee := range following {
		user := newUserNode(*followee.Login, node)
		if user == nil {
			continue
		}

		if user.Login == target {
			done <- user
			return
		}

		jobQueue <- jobRequest{User: user}
	}

	if len(following) == 100 {
		node.Page++
		jobQueue <- jobRequest{User: node}
	}
}

// Run starts the dispatcher and pushes a new request for the root user onto
// the queue. Returns the *UserNode that is received on the done channel.
func Run(start, end string, nWorkers int, g *Github) *UserNode {
	ghClient, ctx = g.Client, context.Background()

	if nWorkers <= 0 {
		nWorkers = 6
	}
	startDispatcher(nWorkers)

	origin, target = start, end
	jobQueue <- jobRequest{User: newUserNode(origin, nil)}

	for {
		select {
		case user := <-done:
			return user
		}
	}
}
