package githubapi

import "github.com/golang/mock/gomock"

func NewMock(ctrl *gomock.Controller) *Client {
	return &Client{
		Apps:         NewMockAppsService(ctrl),
		PullRequests: NewMockPullRequestService(ctrl),
		Repositories: NewMockRepositoriesService(ctrl),
		Users:        NewMockUsersService(ctrl),
	}
}
