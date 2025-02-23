// Copyright 2019 George Aristy
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package commits defines the representations of git commit objects.
package commits

import (
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/vinetiworks/go-gitlint/internal/repo"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// Commits returns commits.
type Commits func() []*Commit

// Commit holds data for a single git commit.
type Commit struct {
	Hash       string
	Message    string
	Date       time.Time
	NumParents int
	Author     *Author
}

// Author is the author of a commit.
type Author struct {
	Name  string
	Email string
}

// ID is the commit's hash.
func (c *Commit) ID() string {
	return c.Hash
}

// ShortID returns the commit hash's short form.
func (c *Commit) ShortID() string {
	return c.Hash[:7]
}

// Subject is the commit message's subject line.
func (c *Commit) Subject() string {
	return strings.Split(c.Message, "\n")[0]
}

// Body is the commit message's body.
func (c *Commit) Body() string {
	body := ""
	parts := strings.Split(c.Message, "\n\n")

	if len(parts) > 1 {
		body = strings.Join(parts[1:], "")
	}

	return body
}

// In returns the commits in the repo.
func In(repository repo.Repo) Commits {
	return func() []*Commit {
		r := repository()

		ref, err := r.Head()
		if err != nil {
			panic(err)
		}

		iter, err := r.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			panic(err)
		}

		commits := make([]*Commit, 0)

		err = iter.ForEach(func(c *object.Commit) error {
			commits = append(
				commits,
				&Commit{
					Hash:       c.Hash.String(),
					Message:    c.Message,
					Date:       c.Author.When,
					NumParents: len(c.ParentHashes),
					Author: &Author{
						Name:  c.Author.Name,
						Email: c.Author.Email,
					},
				},
			)

			return nil
		})
		if err != nil {
			panic(err)
		}

		return commits
	}
}

// Since returns commits authored since time t (format: yyyy-MM-dd).
func Since(t string, cmts Commits) Commits {
	return filtered(
		func(c *Commit) bool {
			start, err := time.Parse("2006-01-02", t)
			if err != nil {
				panic(err)
			}
			return !c.Date.Before(start)
		},
		cmts,
	)
}

// NotAuthoredByNames filters out commits with authors whose names match any of the given patterns.
func NotAuthoredByNames(patterns []string, cmts Commits) Commits {
	return filtered(
		func(c *Commit) bool {
			for _, p := range patterns {
				match, err := regexp.MatchString(p, c.Author.Name)
				if err != nil {
					panic(err)
				}
				if match {
					return false
				}
			}
			return true
		},
		cmts,
	)
}

// NotAuthoredByEmails filters out commits with authors whose emails match any
// of the given patterns.
func NotAuthoredByEmails(patterns []string, cmts Commits) Commits {
	return filtered(
		func(c *Commit) bool {
			for _, p := range patterns {
				match, err := regexp.MatchString(p, c.Author.Email)
				if err != nil {
					panic(err)
				}
				if match {
					return false
				}
			}
			return true
		},
		cmts,
	)
}

// WithMaxParents returns commits that have at most n number of parents.
// Useful for excluding merge commits.
func WithMaxParents(n int, cmts Commits) Commits {
	return filtered(
		func(c *Commit) bool {
			return c.NumParents <= n
		},
		cmts,
	)
}

// MsgIn returns a single fake commit with the message read from this reader.
// This fake commit will have a fake hash and its timestamp will be time.Now().
func MsgIn(reader io.Reader) Commits {
	return func() []*Commit {
		b, err := ioutil.ReadAll(reader)
		if err != nil {
			panic(err)
		}

		return []*Commit{{
			Hash:    "fakehsh",
			Message: string(b),
			Date:    time.Now(),
		}}
	}
}

func filtered(filter func(*Commit) bool, in Commits) (out Commits) {
	return func() []*Commit {
		f := make([]*Commit, 0)

		for _, c := range in() {
			if filter(c) {
				f = append(f, c)
			}
		}

		return f
	}
}
