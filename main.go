package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"
)

// these are determined through trial and error with what youtube accepts.
// they were incredibly hard to figure out, since youtube doesn't properly follow the spec.
var positions = map[string]string{
	"1": " align:left position:0% size:60%",
	"2": "", // default is align:center
	"3": " align:right position:100% size:60%",
	"4": " align:left position:0% line:50% size:60%",
	"5": " position:50% line:50%", // default is align:center
	"6": " align:right position:100% line:50% size:60%",
	"7": " align:left position:0% line:0% size:60%",
	"8": " line:0%", // default is align:center
	"9": " align:right position:100% line:0% size:60%",
}

var replacer = strings.NewReplacer(
	`\N`, "\n",
	`{\i1}`, "<i>", `{\i0}`, "</i>",
	`{\b1}`, "<b>", `{\b0}`, "</b>",
	`{\an7}`, "", `{\an8}`, "", `{\an9}`, "",
	`{\an4}`, "", `{\an5}`, "", `{\an6}`, "",
	`{\an1}`, "", `{\an2}`, "", `{\an3}`, "",
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./simplevtt FILE")
		os.Exit(1)
	}

	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(bytes.NewReader(b))
	scanner.Split(bufio.ScanLines)

	dialogue := make([]*Dialogue, 0)

	for scanner.Scan() {
		line := scanner.Text()

		// we only care about dialogue
		if !strings.HasPrefix(line, "Dialogue:") {
			continue
		}

		d := parseLine(line)
		if d.Text == "" {
			continue
		}

		dialogue = append(dialogue, d)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error scanning .ass file: %v\n", err)
		os.Exit(1)
	}

	// now we're gonna sort by start time.
	sort.Slice(dialogue, func(i, j int) bool {
		startI := dialogue[i].StartTime()
		startJ := dialogue[j].StartTime()

		return startI.Before(startJ)
	})

	// we're outputting to stdout
	fmt.Printf("WEBVTT\n\n")

	for i, d := range dialogue {
		fmt.Print(d.MarshalVTT(i + 1))
	}
}

// Dialogue is a struct holding information about an aegisub line.
// Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
type Dialogue struct {
	Start string
	End   string
	Style string
	Name  string
	Text  string
}

func (d *Dialogue) StartTime() time.Time {
	t, _ := time.Parse("3:04:05.00", d.Start)
	return t
}

func (d *Dialogue) EndTime() time.Time {
	t, _ := time.Parse("3:04:05.00", d.End)
	return t
}

func (d *Dialogue) FormatStart() string {
	return d.StartTime().Format("15:04:05.000")
}

func (d *Dialogue) FormatEnd() string {
	return d.EndTime().Format("15:04:05.000")
}

func (d *Dialogue) Position() string {
	if len(d.Text) > 5 && strings.HasPrefix(d.Text, `{\an`) {
		return positions[string(d.Text[4])]
	}

	return ""
}

/*
MarshalVTT takes a Dialogue line and turns it into WebVTT format.

	INDEX
	START --> END[ POSITION]
	TEXT
	\n
*/
func (d *Dialogue) MarshalVTT(index int) string {

	return fmt.Sprintf("%d\n%s --> %s%s\n%s\n\n",
		index, d.FormatStart(), d.FormatEnd(), d.Position(), replacer.Replace(d.Text),
	)
}

func parseLine(text string) *Dialogue {
	text = strings.TrimPrefix(text, "Dialogue: ")
	sections := strings.SplitN(text, ",", 10)

	return &Dialogue{
		Start: sections[1],
		End:   sections[2],
		Style: sections[3],
		Name:  sections[4],
		Text:  sections[9],
	}
}
