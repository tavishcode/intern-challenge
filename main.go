package main

import (
	"context"
	"fmt"
	"os"
	"encoding/csv"
	"strings"
	"log"
	"io"

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
	if len(os.Args) < 2 {
		log.Fatal("Missing filename argument\n")
	}
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	
	// ensure file is closed at end of program
	defer file.Close() 

	// csv parsing
	r := csv.NewReader(file)

	// skip header line
	line, err := r.Read()

	for {
		// read one line
		line, err = r.Read()
		// reached end of file ?
        if err == io.EOF { 
            break
		}
		// skip entry if error found during read
		if err != nil {
			fmt.Printf(err.Error())
			continue
		}

		// trim leading and trailing whitespace
		line[0] = strings.TrimSpace(line[0])
		line[1] = strings.TrimSpace(line[1])

		// split repositoy field into owner and name components
		repo := strings.Split(line[0],"/")

		// skip repository if name is incorrectly formmated
		if len(repo) < 2 {
			fmt.Printf("Repository field incorrectly formatted '%s'\n", line[0])
			continue
		}

		// retrieve releases from Github
		releases, _, err := client.Repositories.ListReleases(ctx, repo[0], repo[1], opt)

		// skip repository if releases could not be retrieved
		if err != nil {
			fmt.Printf("Could not retrieve releases for '%s'\n", line[0])
			continue
		}

		minVersion, err := semver.NewVersion(line[1])
		// skip entry if min version is not in valid format
		if err != nil {
			fmt.Printf("Minversion '%s' is invalid\n", line[1])
			continue
		}

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

		// get latest versions
		versionSlice := LatestVersions(allReleases, minVersion)
		if versionSlice != nil {
			fmt.Printf("latest versions of %s: %s\n", line[0], versionSlice)
		} else {
			fmt.Printf("Min version is greater than latest available version\n")
		}
    }
}
