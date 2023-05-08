package main

import (
	"flag"
	"fmt"
	"github.com/karrick/godirwalk"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var pth string
var targetDur time.Duration
var before string
var dryrun bool

func main() {

	flag.StringVar(&pth, "path", "", "The Path to search")
	flag.StringVar(&before, "before", "", "Remove before duration ie 120h = Remove anything from 120 hours ago or before")
	flag.BoolVar(&dryrun, "dryrun", false, "Dry run on delete")
	flag.Parse()

	dur, err := time.ParseDuration(strings.Trim(before, " "))

	if err == nil {

		targetDur = dur

	} else {

		fmt.Println(err)
		targetDur = time.Duration(time.Hour * 120)

	}

	fmt.Println("Searching Path: ", pth)

	if len(pth) < 5 {
		fmt.Println("No Path, or Path too short")
		return
	}

	filepath.WalkDir(pth, walk)

	pruneEmptyDirectories(pth)

}

func pruneEmptyDirectories(osDirname string) (int, error) {
	var count int

	err := godirwalk.Walk(osDirname, &godirwalk.Options{
		Unsorted: true,
		Callback: func(_ string, _ *godirwalk.Dirent) error {
			// no-op while diving in; all the fun happens in PostChildrenCallback
			return nil
		},
		PostChildrenCallback: func(osPathname string, _ *godirwalk.Dirent) error {
			s, err := godirwalk.NewScanner(osPathname)
			if err != nil {
				return err
			}

			// Attempt to read only the first directory entry. Remember that
			// Scan skips both "." and ".." entries.
			hasAtLeastOneChild := s.Scan()

			// If error reading from directory, wrap up and return.
			if err := s.Err(); err != nil {
				return err
			}

			if hasAtLeastOneChild {
				return nil // do not remove directory with at least one child
			}
			if osPathname == osDirname {
				return nil // do not remove directory that was provided top-level directory
			}

			err = os.Remove(osPathname)
			if err == nil {
				count++
			}
			return err
		},
	})

	return count, err
}

func walk(s string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if !d.IsDir() {

		file, err := os.Stat(s)

		if err != nil {
			fmt.Println(err)
		}

		modifiedtime := file.ModTime()

		nw := time.Now()

		diff := nw.Sub(modifiedtime)

		if diff > targetDur {

			go func() {
				
				if dryrun == false {
					fmt.Printf("Remove %s created @ %s\n", s, modifiedtime)
					re := os.Remove(s)
					if re != nil {
						fmt.Println(re)
					}
				} else {

					fmt.Printf("[DRY RUN] Remove %s created @ %s\n", s, modifiedtime)
				}

			}()

		}

	}
	return nil
}
