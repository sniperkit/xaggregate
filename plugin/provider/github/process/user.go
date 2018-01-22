package github

import (
	"context"

	"github.com/google/go-github/github"

	"github.com/sniperkit/xtask/plugin/aggregate/service/github/model"

	gitClient "github.com/fengyfei/nuts/github/client"
)

// GetOwnerByID 调用 GitHub API 获取作者信息
func GetOwnerByID(ownerID int, client *gitClient.GHClient) (*models.MDUser, *github.Response, error) {
	owner, resp, err := client.Client.Users.GetByID(context.Background(), ownerID)
	if err != nil {
		if resp != nil {
			return nil, resp, err
		}

		return nil, nil, err
	}

	user := &models.MDUser{
		Login:             owner.Login,
		ID:                owner.ID,
		HTMLURL:           owner.HTMLURL,
		Name:              owner.Name,
		Email:             owner.Email,
		PublicRepos:       owner.PublicRepos,
		PublicGists:       owner.PublicGists,
		Followers:         owner.Followers,
		Following:         owner.Following,
		CreatedAt:         owner.CreatedAt,
		UpdatedAt:         owner.UpdatedAt,
		SuspendedAt:       owner.SuspendedAt,
		Type:              owner.Type,
		TotalPrivateRepos: owner.TotalPrivateRepos,
		OwnedPrivateRepos: owner.OwnedPrivateRepos,
		PrivateGists:      owner.PrivateGists,
	}

	return user, resp, nil
}
