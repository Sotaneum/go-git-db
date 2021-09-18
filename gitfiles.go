package gitfiles

import (
	"io/ioutil"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	file "github.com/Sotaneum/go-json-file"
	ktime "github.com/Sotaneum/go-kst-time"
)

type User struct {
	Name  string
	Email string
}

// GitFiles : git 객체입니다.
type GitFiles struct {
	path       string
	url        string
	token      string
	remoteName string
	user       User
}

// Clear : 원격 저장소간에 문제가 발생했을 경우 폴더를 비우고 재실행합니다.
func (gitFiles *GitFiles) Clear() error {
	files, err := ioutil.ReadDir(gitFiles.path)

	if err != nil {
		return err
	}

	for _, f := range files {
		f := file.File{Path: gitFiles.path, Name: f.Name()}
		f.Remove()
	}

	return nil
}

func (gitFiles *GitFiles) getAuth() transport.AuthMethod {
	return &http.BasicAuth{
		Password: gitFiles.token,
		Username: gitFiles.user.Name,
	}
}

func (gitFiles *GitFiles) getAuthor() *object.Signature {
	return &object.Signature{
		When:  ktime.GetNow(),
		Name:  gitFiles.user.Name,
		Email: gitFiles.user.Email,
	}
}

func (gitFiles *GitFiles) getRepository() (*git.Repository, error) {
	localRep, localRepErr := git.PlainOpen(gitFiles.path)

	if localRepErr == nil {
		return localRep, nil
	}

	return git.PlainClone(gitFiles.path, false, &git.CloneOptions{
		URL:      gitFiles.url,
		Progress: os.Stdout,
		Auth:     gitFiles.getAuth(),
	})
}

func (gitFiles *GitFiles) getWorktree() (*git.Worktree, *git.Repository, error) {
	rep, repErr := gitFiles.getRepository()

	if repErr != nil {
		return nil, nil, repErr
	}

	work, workErr := rep.Worktree()

	return work, rep, workErr
}

// Init : Git 레포를 가져오거나 기존 것을 불러옵니다.
func (gitFiles *GitFiles) Init(user User, path, repURL, accessToken, remoteName string) error {
	gitFiles.path = path
	gitFiles.url = repURL
	gitFiles.token = accessToken
	gitFiles.remoteName = remoteName
	gitFiles.user = user

	err := gitFiles.Pull()

	if err == nil {
		return nil
	}

	if err == git.NoErrAlreadyUpToDate {
		return nil
	}

	gitFiles.Clear()
	clearErr := gitFiles.Pull()

	if clearErr == git.NoErrAlreadyUpToDate {
		return nil
	}

	return clearErr
}

// Pull : Git 레포 정보를 가져옵니다.
func (gitFiles *GitFiles) Pull() error {
	work, _, err := gitFiles.getWorktree()

	if err != nil {
		return err
	}

	return work.Pull(&git.PullOptions{RemoteName: gitFiles.remoteName, Auth: gitFiles.getAuth()})
}

// Push : Add하고 Commit하고 Push 합니다.
func (gitFiles *GitFiles) Push() error {
	work, rep, err := gitFiles.getWorktree()

	if err != nil {
		return err
	}

	if addErr := work.AddGlob("."); addErr != nil {
		return addErr
	}

	now := ktime.GetNow().Format("2006-01-02T15:04:05")

	if _, commitErr := work.Commit(now, &git.CommitOptions{All: true, Author: gitFiles.getAuthor()}); commitErr != nil {
		return commitErr
	}

	return rep.Push(&git.PushOptions{Force: true, RemoteName: gitFiles.remoteName, Auth: gitFiles.getAuth(), Progress: os.Stdout})
}
