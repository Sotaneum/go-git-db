package gitfiles_test

import (
	"testing"

	gitfiles "github.com/Sotaneum/go-git-files"
	file "github.com/Sotaneum/go-json-file"
)

func TestRunner(t *testing.T) {
	user := &gitfiles.User{Name: "Sotaneum", Email: "test@a.b.c"}
	files := &gitfiles.GitFiles{}
	files.Init(*user, "./repo", "https://github.com/Sotaneum/go-git-files-test", "", "origin")
	f := &file.File{Path: "./repo", Name: "new_file.json"}
	f.Save("테스트입니다!")
	files.Push()
}
