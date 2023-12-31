// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/go-github/v53/github"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
)

const codeownersHeader = `# Code generated by githubgen. DO NOT EDIT.
#####################################################
#
# List of approvers for OpenTelemetry Collector Contrib
#
#####################################################
#
# Learn about membership in OpenTelemetry community:
# https://github.com/open-telemetry/community/blob/main/community-membership.md
#
#
# Learn about CODEOWNERS file format:
# https://help.github.com/en/articles/about-code-owners
#
# NOTE: Lines should be entered in the following format:
# <component_path_relative_from_project_root>/<min_1_space><owner_1><space><owner_2><space>..<owner_n>
# extension/oauth2clientauthextension/                 @open-telemetry/collector-contrib-approvers @pavankrish123 @jpkrohling
# Path separator and minimum of 1 space between component path and owners is
# important for validation steps
#

* @open-telemetry/collector-contrib-approvers
`

const unmaintainedHeader = `

## UNMAINTAINED components
## The Github issue template generation code needs this to generate the corresponding labels.

`

const allowlistHeader = `# Code generated by githubgen. DO NOT EDIT.
#####################################################
#
# List of components in OpenTelemetry Collector Contrib
# waiting on owners to be assigned
#
#####################################################
#
# Learn about membership in OpenTelemetry community:
#  https://github.com/open-telemetry/community/blob/main/community-membership.md
#
#
# Learn about CODEOWNERS file format:
#  https://help.github.com/en/articles/about-code-owners
#

## 
# NOTE: New components MUST have a codeowner. Add new components to the CODEOWNERS file instead of here.
##

## COMMON & SHARED components
internal/common

`

const unmaintainedStatus = "unmaintained"

// Generates files specific to Github according to status metadata:
// .github/CODEOWNERS
// .github/ALLOWLIST
func main() {
	folder := flag.String("folder", ".", "folder investigated for codeowners")
	allowlistFilePath := flag.String("allowlist", "cmd/githubgen/allowlist.txt", "path to a file containing an allowlist of members outside the OpenTelemetry organization")
	flag.Parse()
	if err := run(*folder, *allowlistFilePath); err != nil {
		log.Fatal(err)
	}
}

type Codeowners struct {
	// Active codeowners
	Active []string `mapstructure:"active"`
	// Emeritus codeowners
	Emeritus []string `mapstructure:"emeritus"`
}
type Status struct {
	Stability     map[string][]string `mapstructure:"stability"`
	Distributions []string            `mapstructure:"distributions"`
	Class         string              `mapstructure:"class"`
	Warnings      []string            `mapstructure:"warnings"`
	Codeowners    *Codeowners         `mapstructure:"codeowners"`
}
type metadata struct {
	// Type of the component.
	Type string `mapstructure:"type"`
	// Type of the parent component (applicable to subcomponents).
	Parent string `mapstructure:"parent"`
	// Status information for the component.
	Status *Status `mapstructure:"status"`
}

func loadMetadata(filePath string) (metadata, error) {
	cp, err := fileprovider.New().Retrieve(context.Background(), "file:"+filePath, nil)
	if err != nil {
		return metadata{}, err
	}

	conf, err := cp.AsConf()
	if err != nil {
		return metadata{}, err
	}

	md := metadata{}
	if err := conf.Unmarshal(&md); err != nil {
		return md, err
	}

	return md, nil
}

func run(folder string, allowlistFilePath string) error {
	members, err := getGithubMembers()
	if err != nil {
		return err
	}
	allowlistData, err := os.ReadFile(allowlistFilePath)
	if err != nil {
		return err
	}
	allowlistLines := strings.Split(string(allowlistData), "\n")

	allowlist := make(map[string]struct{}, len(allowlistLines))
	for _, line := range allowlistLines {
		allowlist[line] = struct{}{}
	}

	components := map[string]metadata{}
	var foldersList []string
	maxLength := 0
	allCodeowners := map[string]struct{}{}
	err = filepath.Walk(folder, func(path string, info fs.FileInfo, err error) error {
		if info.Name() == "metadata.yaml" {
			m, err := loadMetadata(path)
			if err != nil {
				return err
			}
			if m.Status == nil {
				return nil
			}
			key := filepath.Dir(path) + "/"
			components[key] = m
			foldersList = append(foldersList, key)
			for stability := range m.Status.Stability {
				if stability == unmaintainedStatus {
					// do not account for unmaintained status to change the max length of the component line.
					return nil
				}
			}
			for _, id := range m.Status.Codeowners.Active {
				allCodeowners[id] = struct{}{}
			}
			if len(key) > maxLength {
				maxLength = len(key)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	sort.Strings(foldersList)
	var missingCodeowners []string
	var duplicateCodeowners []string
	for codeowner := range allCodeowners {
		_, present := members[codeowner]

		if !present {
			_, allowed := allowlist[codeowner]
			allowed = allowed || strings.HasPrefix(codeowner, "open-telemetry/")
			if !allowed {
				missingCodeowners = append(missingCodeowners, codeowner)
			}
		} else if _, ok := allowlist[codeowner]; ok {
			duplicateCodeowners = append(duplicateCodeowners, codeowner)
		}
	}
	if len(missingCodeowners) > 0 {
		sort.Strings(missingCodeowners)
		return fmt.Errorf("codeowners are not members: %s", strings.Join(missingCodeowners, ", "))
	}
	if len(duplicateCodeowners) > 0 {
		sort.Strings(duplicateCodeowners)
		return fmt.Errorf("codeowners members duplicate in allowlist: %s", strings.Join(duplicateCodeowners, ", "))
	}

	codeowners := codeownersHeader
	deprecatedList := "## DEPRECATED components\n"
	unmaintainedList := "\n## UNMAINTAINED components\n"

	unmaintainedCodeowners := unmaintainedHeader
	currentFirstSegment := ""
LOOP:
	for _, key := range foldersList {
		m := components[key]
		for stability := range m.Status.Stability {
			if stability == unmaintainedStatus {
				unmaintainedList += key + "\n"
				unmaintainedCodeowners += fmt.Sprintf("%s%s @open-telemetry/collector-contrib-approvers \n", key, strings.Repeat(" ", maxLength-len(key)))
				continue LOOP
			}
			if stability == "deprecated" && (m.Status.Codeowners == nil || len(m.Status.Codeowners.Active) == 0) {
				deprecatedList += key + "\n"
			}
		}

		if m.Status.Codeowners != nil {
			parts := strings.Split(key, string(os.PathSeparator))
			firstSegment := parts[0]
			if firstSegment != currentFirstSegment {
				currentFirstSegment = firstSegment
				codeowners += "\n"
			}
			owners := ""
			for _, owner := range m.Status.Codeowners.Active {
				owners += " "
				owners += "@" + owner
			}
			codeowners += fmt.Sprintf("%s%s @open-telemetry/collector-contrib-approvers%s\n", key, strings.Repeat(" ", maxLength-len(key)), owners)
		}
	}

	err = os.WriteFile(filepath.Join(".github", "CODEOWNERS"), []byte(codeowners+unmaintainedCodeowners), 0600)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(".github", "ALLOWLIST"), []byte(allowlistHeader+deprecatedList+unmaintainedList), 0600)
	if err != nil {
		return err
	}

	return nil
}

func getGithubMembers() (map[string]struct{}, error) {
	client := github.NewTokenClient(context.Background(), os.Getenv("GITHUB_TOKEN"))
	var allUsers []*github.User
	pageIndex := 0
	for {
		users, resp, err := client.Organizations.ListMembers(context.Background(), "open-telemetry",
			&github.ListMembersOptions{
				PublicOnly: false,
				ListOptions: github.ListOptions{
					PerPage: 50,
					Page:    pageIndex,
				},
			},
		)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if len(users) == 0 {
			break
		}
		allUsers = append(allUsers, users...)
		pageIndex++
	}

	usernames := make(map[string]struct{}, len(allUsers))
	for _, u := range allUsers {
		usernames[*u.Login] = struct{}{}
	}
	return usernames, nil
}
