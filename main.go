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
	semver.Sort(releases) // sort releases in ascending order
	for i := len(releases)-1; i >= 0; i-- { // iterate through slice backwards
		if !releases[i].LessThan(*minVersion) {
			if i == len(releases)-1 || releases[i].Minor != versionSlice[len(versionSlice)-1].Minor {
				versionSlice =  append(versionSlice, releases[i])
			}
		}
	}
	return versionSlice
}

// Here we implement the basics of communicating with github through the library as well as printing the version
// You will need to implement LatestVersions function as well as make this application support the file format outlined in the README
// Please use the format defined by the fmt.Printf line at the bottom, as we will define a passing coding challenge as one that outputs
// the correct information, including this line
func main() {
	// Github
	client := github.NewClient(nil) // ':=' serves as both declare and init
	ctx := context.Background()
	opt := &github.ListOptions{PerPage: 10}
	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	r := csv.NewReader(file)
	lines, err := r.ReadAll()
	if err != nil {
		panic( err)
	}
	for i, line := range lines {
		if i == 0 {
			// skip header line
			continue
		}
		repo := strings.Split(line[0],"/")
		releases, _, err := client.Repositories.ListReleases(ctx, repo[0], repo[1], opt)
		if err != nil {
			fmt.Printf("Could not retrieve releases for '%s'\n", line[0])
			continue
		}
		minVersion := semver.New(line[1])
		allReleases := make([]*semver.Version, len(releases))
		for i, release := range releases {
			versionString := *release.TagName
			if versionString[0] == 'v' {
				versionString = versionString[1:]
			}
			allReleases[i] = semver.New(versionString)
		}
		versionSlice := LatestVersions(allReleases, minVersion)
		fmt.Printf("latest versions of %s: %s\n", line[0], versionSlice)
	}
}
