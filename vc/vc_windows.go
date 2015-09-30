// +build windows

package vc

import (
	"fmt"
	"time"

	git "github.com/libgit2/git2go"
)

func Commit(path string, message string) error {

	repo, err := git.OpenRepository(".gattai")

	// TODO re-config
	signature := &git.Signature{
		Name:  "Chanwit Kaewkasi",
		Email: "chanwit@gmail.com",
		When:  time.Now(),
	}

	idx, err := repo.Index()
	if err != nil {
		panic(err)
	}

	err = idx.AddByPath(path)
	if err != nil {
		panic(err)
	}

	treeId, err := idx.WriteTree()
	if err != nil {
		panic(err)
	}

	err = idx.Write()
	if err != nil {
		panic(err)
	}

	tree, err := repo.LookupTree(treeId)
	if err != nil {
		panic(err)
	}

	var commit *git.Oid
	head, err := repo.Head()
	if err == nil {
		headCommit, err := repo.LookupCommit(head.Target())
		if err != nil {
			panic(err)
		}

		commit, err = repo.CreateCommit("HEAD", signature, signature, message, tree, headCommit)
		if err != nil {
			panic(err)
		}

	} else {
		commit, err = repo.CreateCommit("HEAD", signature, signature, message, tree)
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("%s\n", commit)
	return err
}
