package process

import (
	gitClient "github.com/fengyfei/nuts/github/client"
)

// 填入自己生成的 token
var tokens []string = []string{}

var clientManager *gitClient.ClientManager = gitClient.NewManager(tokens)
