// tgrep.go
//  tgrep read stdin and find N lines within best timestamp distance
//
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
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
)

type keep struct {
	text     string
	distance float64
}

func main() {

	flag.StringVar(&stime, "t", stampRef, "approximate timestamp to look after")
	flag.IntVar(&lines, "n", 1, "#lines to ouput")
	flag.BoolVar(&version, "v", false, "prints current version")
	flag.Parse()

	if version {
		fmt.Println("tgrep v0.22")
		os.Exit(0)
	}
	if lines < 1 {
		log.Fatal("lines must be >0  ")
	}
	t, err := time.Parse(time.Stamp, stime)
	if err != nil {
		log.Fatal("ERROR: parsing timestamp ", err)
	}

	stampRefT, _ := time.Parse(time.Stamp, stampRef)
	bestdistance := math.Abs(float64(t.Sub(stampRefT))) // far distance

	// buffer for keeping best lines
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
		if results != nil {
			d, err := time.Parse(time.Stamp, results["date"])
			if err != nil {
				log.Printf("Error in date parsing -->%s<-- \n", results["date"])
				continue
			}
			distance := math.Abs(float64(t.Sub(d)))

			if (distance <= keeps[lines-1].distance) || (distance <= keeps[0].distance) {
				for line := 1; line < lines; line++ {
					if keeps[line].distance != 0 { // avoid copy if unitialised
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
	}

	for line := 0; line < lines; line++ {
		if keeps[line].text != "" { // keeps may contain empty lines
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
