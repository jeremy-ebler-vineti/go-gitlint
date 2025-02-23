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

package issues_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vinetiworks/go-gitlint/internal/commits"
	"github.com/vinetiworks/go-gitlint/internal/issues"
)

func TestCollected(t *testing.T) {
	expected := []*commits.Commit{
		{Hash: "123"},
		{Hash: "456"},
	}
	isus := issues.Collected(
		[]issues.Filter{
			func(c *commits.Commit) issues.Issue {
				var issue issues.Issue

				if c.Hash == "123" || c.Hash == "456" {
					issue = issues.Issue{
						Desc:   "test",
						Commit: *c,
					}
				}

				return issue
			},
		},
		func() []*commits.Commit {
			return append(expected, &commits.Commit{Hash: "789"})
		},
	)()

	assert.Len(t,
		isus,
		2,
		"issues.Collected() must return the filtered commits")

	for _, i := range isus {
		assert.Contains(t,
			expected, &i.Commit,
			"issues.Collected() must return the filtered commits")
	}
}

func TestPrinted(t *testing.T) {
	const sep = "-"

	isus := []issues.Issue{
		{
			Desc: "issueA",
			Commit: commits.Commit{
				Hash:    "18045269d8d2fd8f53d01883c6c7b548d0b9e3ae",
				Message: "first commit",
			},
		},
		{
			Desc: "issueB",
			Commit: commits.Commit{
				Hash:    "4be918ff8bfc91de77a1462707a8d2eb30956f93",
				Message: "second commit",
			},
		},
	}

	var expected string

	for _, i := range isus {
		expected += fmt.Sprintf("%s: %s%s", i.Commit.ShortID(), i.Desc, sep)
	}

	writer := &mockWriter{}

	issues.Printed(
		writer, sep,
		func() []issues.Issue {
			return isus
		},
	)()

	assert.Equal(t,
		expected, writer.msg,
		"issues.Printed() must join Commit.ShortID() and the Issue.Desc with the separator")
}

type mockWriter struct {
	msg string
}

func (m *mockWriter) Write(b []byte) (int, error) {
	m.msg += string(b)
	return len(b), nil
}
