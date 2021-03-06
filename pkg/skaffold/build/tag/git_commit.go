/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tag

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"

	"github.com/GoogleCloudPlatform/skaffold/pkg/skaffold/util"
	"github.com/pkg/errors"
)

// GitCommit tags an image by the git commit it was built at.
type GitCommit struct {
}

// GenerateFullyQualifiedImageName tags an image with the supplied image name and the git commit.
func (c *GitCommit) GenerateFullyQualifiedImageName(opts *TagOptions) (string, error) {
	// If the repository state is dirty, we add a -dirty-unique-id suffix to work well with local iterations
	dirtyCmd := exec.Command("git", "status", "--porcelain")
	stdout, _, err := util.RunCommand(dirtyCmd, nil)
	if err != nil {
		return "", errors.Wrap(err, "determining repo state")
	}
	suffix := ""
	if string(stdout) != "" {
		// The file state is dirty. To generate a unique suffix, let's hash the "git diff" output.
		// It should be roughly content-addressable.
		uniqueCmd := exec.Command("git", "diff")
		stdout, _, err := util.RunCommand(uniqueCmd, nil)
		if err != nil {
			return "", errors.Wrap(err, "determining git diff")
		}
		sha := sha256.Sum256(stdout)
		shaStr := hex.EncodeToString(sha[:])[:16]
		suffix = fmt.Sprintf("dirty-%s", shaStr)
	}
	cmd := exec.Command("git", "rev-parse", "HEAD")
	stdout, _, err = util.RunCommand(cmd, nil)
	if err != nil {
		return "", errors.Wrap(err, "determining current git commit")
	}
	commit := strings.TrimSuffix(string(stdout), "\n")
	if suffix != "" {
		return fmt.Sprintf("%s:%s-%s", opts.ImageName, commit, suffix), nil
	}
	return fmt.Sprintf("%s:%s", opts.ImageName, commit), nil
}
