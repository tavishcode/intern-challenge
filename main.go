package main

import (
	"context"
	"fmt"
	"os"
	"encoding/csv"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/google/go-github/github"
)

// LatestVersions returns a sorted slice with the highest version as its first element and the highest version of the smaller minor versions in a descending order
func LatestVersions(releases []*semver.Version, minVersion *semver.Version) []*semver.Version {
	var versionSlice []*semver.Version
	// sort releases in ascending order
	semver.Sort(releases) 
	// iterate through slice backwards
	for i := len(releases)-1; i >= 0; i-- {
		// release > min version ?
		if !releases[i].LessThan(*minVersion) {
			// if first element or element has diff minor version than last item in versionSlice,
			// then add to versionSlice
			if i == len(releases)-1 || releases[i].Minor != versionSlice[len(versionSlice)-1].Minor {
				versionSlice =  append(versionSlice, releases[i])
			}
		}
	}
	return versionSlice
}

func main() {
	// init Github api reqs
	client := github.NewClient(nil)
	ctx := context.Background()
	opt := &github.ListOptions{PerPage: 10}

	// get filename from cmd args
	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	// csv parsing
	r := csv.NewReader(file)
	lines, err := r.ReadAll()
	if err != nil {
		panic( err)
	}

	for i, line := range lines {
		// skip header line
		if i == 0 {
			continue
		}

		// split repositoy field into owner and name components
		repo := strings.Split(line[0],"/")

		releases, _, err := client.Repositories.ListReleases(ctx, repo[0], repo[1], opt)

		// skip repository if releases could not be retrieved
		if err != nil {
			fmt.Printf("Could not retrieve releases for '%s'\n", line[0])
			continue
		}

		minVersion := semver.New(line[1])
		allReleases := make([]*semver.Version, len(releases))

		for i, release := range releases {
			versionString := *release.TagName
			
			// remove preceeding 'v' from version strings
			if versionString[0] == 'v' {
				versionString = versionString[1:]
			}

			// store version string as a semver struct
			allReleases[i] = semver.New(versionString)
		}
		versionSlice := LatestVersions(allReleases, minVersion)
		fmt.Printf("latest versions of %s: %s\n", line[0], versionSlice)
	}
}
