// tgrep.go
//  tgrep read stdin and find N lines within best timestamp distance
//
// Author : thc2cat@gmail.com
// 2018, ?? License
// 2021 v0.23 Bug corrections
//

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"
)

var (
	stime   string
	lines   int
	version bool

	stampRegexp = regexp.MustCompile(`^(?P<date>(... .. ..:..:..))`)
	stampRef    = "Jan 1 00:00:00"
	versionS    = "tgrep v0.23"
)

type keep struct {
	text     string
	distance int64
}

func main() {

	flag.StringVar(&stime, "t", stampRef, "approximate timestamp to look after")
	flag.IntVar(&lines, "n", 1, "#lines to output")
	flag.BoolVar(&version, "v", false, fmt.Sprintf("prints current version :  %s", versionS))
	flag.Parse()

	if version {
		fmt.Println(versionS)
		os.Exit(0)
	}
	if lines < 1 {
		log.Fatal("ERROR: line count must be >0 ")
	}
	t, err := time.Parse(time.Stamp, stime)
	if err != nil {
		log.Fatal("ERROR: parsing timestamp ", err)
	}

	stampRefT, _ := time.Parse(time.Stamp, stampRef)
	bestdistance := WithTwosComplement(int64(t.Sub(stampRefT)))

	// buffer for keeping nearest lines
	keeps := make([]keep, lines)
	keeps[0].distance = bestdistance
	keeps[lines-1].distance = bestdistance

	// f, _ := os.Open("test.log")
	// defer f.Close()
	// s := bufio.NewScanner(f)
	// Scanner read from Stdin otherwise
	s := bufio.NewScanner(os.Stdin)

	for s.Scan() {
		texte := s.Text()
		results := reSubMatchMap(stampRegexp, texte)
		if results == nil {
			continue
		}
		d, err := time.Parse(time.Stamp, results["date"])
		if err != nil {
			log.Printf("Error in date parsing -->%s<-- \n", results["date"])
			continue
		}
		distance := WithTwosComplement(int64(t.Sub(d)))

		if (distance == 0 || distance <= keeps[lines-1].distance) || (distance < keeps[0].distance) {
			for line := 1; line < lines; line++ { // Keep best lines
				if keeps[line].text != "" { // avoid copy if unitialised
					keeps[line-1].distance = keeps[line].distance
					keeps[line-1].text = keeps[line].text
				}
			}
			keeps[lines-1].distance = distance
			keeps[lines-1].text = texte
		}
		if (distance > keeps[lines-1].distance) && (distance > keeps[0].distance) {
			break // too far in the file
		}

	}

	for line := 0; line < lines; line++ {
		if (keeps[line].text != "") || (keeps[line].distance != 0) { // keeps may contain empty lines
			fmt.Println(keeps[line].text)
		}
	}
	os.Exit(0)
}

func reSubMatchMap(r *regexp.Regexp, str string) map[string]string {
	match := r.FindStringSubmatch(str)
	if len(match) == 0 {
		return nil
	}
	subMatchMap := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}
	return subMatchMap
}

// WithTwosComplement return Abs without math.
func WithTwosComplement(n int64) int64 {
	// http://cavaliercoder.com/blog/optimized-abs-for-int64-in-go.html
	y := n >> 63       // y ← x ⟫ 63
	return (n ^ y) - y // (x ⨁ y) - y
}
