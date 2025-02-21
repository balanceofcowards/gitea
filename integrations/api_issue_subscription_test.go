// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"fmt"
	"net/http"
	"testing"

	issues_model "code.gitea.io/gitea/models/issues"
	repo_model "code.gitea.io/gitea/models/repo"
	"code.gitea.io/gitea/models/unittest"
	user_model "code.gitea.io/gitea/models/user"
	api "code.gitea.io/gitea/modules/structs"

	"github.com/stretchr/testify/assert"
)

func TestAPIIssueSubscriptions(t *testing.T) {
	defer prepareTestEnv(t)()

	issue1 := unittest.AssertExistsAndLoadBean(t, &issues_model.Issue{ID: 1})
	issue2 := unittest.AssertExistsAndLoadBean(t, &issues_model.Issue{ID: 2})
	issue3 := unittest.AssertExistsAndLoadBean(t, &issues_model.Issue{ID: 3})
	issue4 := unittest.AssertExistsAndLoadBean(t, &issues_model.Issue{ID: 4})
	issue5 := unittest.AssertExistsAndLoadBean(t, &issues_model.Issue{ID: 8})

	owner := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: issue1.PosterID})

	session := loginUser(t, owner.Name)
	token := getTokenForLoggedInUser(t, session)

	testSubscription := func(issue *issues_model.Issue, isWatching bool) {
		issueRepo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: issue.RepoID})

		urlStr := fmt.Sprintf("/api/v1/repos/%s/%s/issues/%d/subscriptions/check?token=%s", issueRepo.OwnerName, issueRepo.Name, issue.Index, token)
		req := NewRequest(t, "GET", urlStr)
		resp := session.MakeRequest(t, req, http.StatusOK)
		wi := new(api.WatchInfo)
		DecodeJSON(t, resp, wi)

		assert.EqualValues(t, isWatching, wi.Subscribed)
		assert.EqualValues(t, !isWatching, wi.Ignored)
		assert.EqualValues(t, issue.APIURL()+"/subscriptions", wi.URL)
		assert.EqualValues(t, issue.CreatedUnix, wi.CreatedAt.Unix())
		assert.EqualValues(t, issueRepo.APIURL(), wi.RepositoryURL)
	}

	testSubscription(issue1, true)
	testSubscription(issue2, true)
	testSubscription(issue3, true)
	testSubscription(issue4, false)
	testSubscription(issue5, false)

	issue1Repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: issue1.RepoID})
	urlStr := fmt.Sprintf("/api/v1/repos/%s/%s/issues/%d/subscriptions/%s?token=%s", issue1Repo.OwnerName, issue1Repo.Name, issue1.Index, owner.Name, token)
	req := NewRequest(t, "DELETE", urlStr)
	session.MakeRequest(t, req, http.StatusCreated)
	testSubscription(issue1, false)

	req = NewRequest(t, "DELETE", urlStr)
	session.MakeRequest(t, req, http.StatusOK)
	testSubscription(issue1, false)

	issue5Repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: issue5.RepoID})
	urlStr = fmt.Sprintf("/api/v1/repos/%s/%s/issues/%d/subscriptions/%s?token=%s", issue5Repo.OwnerName, issue5Repo.Name, issue5.Index, owner.Name, token)
	req = NewRequest(t, "PUT", urlStr)
	session.MakeRequest(t, req, http.StatusCreated)
	testSubscription(issue5, true)

	req = NewRequest(t, "PUT", urlStr)
	session.MakeRequest(t, req, http.StatusOK)
	testSubscription(issue5, true)
}
