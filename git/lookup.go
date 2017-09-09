package git

import (
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

type refCommit struct {
	ref    string
	commit *object.Commit
}

type lookups []func(*git.Repository, string) ([]*refCommit, error)

func lookup(repo *git.Repository, commitish string) ([]*refCommit, error) {
	funcs := lookups{
		lookupTag,
		lookupRef,
		lookupCommit,
	}

	for _, f := range funcs {
		if commits, err := f(repo, commitish); len(commits) != 0 || err != nil {
			return commits, err
		}
	}

	return nil, nil
}

func lookupRef(repo *git.Repository, ref string) ([]*refCommit, error) {
	refs, err := repo.References()
	if err != nil {
		return nil, err
	}

	ref = "/" + ref
	commits := make([]*refCommit, 0, 1)
	err = refs.ForEach(func(r *plumbing.Reference) error {
		name := r.Name().String()
		if strings.HasPrefix(name, "refs/heads/") ||
			strings.HasPrefix(name, "refs/tags/") {
			return nil
		}

		if strings.HasSuffix(name, ref) {
			hash := r.Hash()
			commit, err := repo.CommitObject(hash)
			if err != nil {
				if err == plumbing.ErrObjectNotFound {
					return nil
				}
				return err
			}

			commits = append(commits, &refCommit{r.Name().String(), commit})
		}

		return nil
	})

	return commits, err
}

func lookupTag(repo *git.Repository, tag string) ([]*refCommit, error) {
	tags, err := repo.TagObjects()
	if err != nil {
		return nil, err
	}

	var commits []*refCommit
	err = tags.ForEach(func(t *object.Tag) error {
		if t.Name != tag {
			return nil
		}

		c, err := t.Commit()
		if err != nil {
			return nil
		}

		commits = []*refCommit{{t.Name, c}}
		return storer.ErrStop
	})

	return commits, err
}

func lookupCommit(repo *git.Repository, commit string) ([]*refCommit, error) {
	if len(commit) < 5 {
		return nil, nil
	}

	if len(commit) == 40 {
		c, err := repo.CommitObject(plumbing.NewHash(commit))
		if err != nil {
			return nil, err
		}

		return []*refCommit{{commit, c}}, nil
	}

	list, err := repo.CommitObjects()
	if err != nil {
		return nil, err
	}

	commits := make([]*refCommit, 0, 1)
	err = list.ForEach(func(c *object.Commit) error {
		h := c.Hash.String()
		if strings.HasPrefix(h, commit) {
			commits = append(commits, &refCommit{h, c})
		}

		return nil
	})

	return commits, err
}
