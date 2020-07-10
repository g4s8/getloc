package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/caarlos0/spin"
	"github.com/google/go-github/v31/github"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

// GitHub client
var client *github.Client

// GitHub network context
var ctx context.Context

func main() {
	ctx = context.Background()
	client = github.NewClient(nil)
	if len(os.Args) < 2 {
		fmt.Errorf("Usage: %s <coordinates>\n"+
			"  - coordinates - repo, user or organization",
			os.Args[0])
		os.Exit(1)
	}
	coords := os.Args[1]
	var repos []*github.Repository
	s := spin.New("Working %s")
	s.Set(spin.Spin1)
	defer s.Stop()
	if strings.Contains(coords, "/") {
		split := strings.Split(coords, "/")
		repo, _, err := client.Repositories.Get(ctx, split[0], split[1])
		if err != nil {
			panic(err)
		}
		repos = []*github.Repository{repo}
		fmt.Printf("Found %s repository\n", repo.GetFullName())
	} else {
		s.Start()
		rps, _, err := client.Repositories.ListByOrg(ctx, coords, nil)
		if err != nil {
			panic(err)
		}
		repos = rps
		s.Stop()
		fmt.Println(spin.ClearLine)
		fmt.Printf("Found %d repos:\n", len(rps))
		for _, r := range rps {
			fmt.Printf(" - %s\n", r.GetFullName())
		}
	}
	s.Start()
	total := new(loc)
	for _, repo := range repos {
		l, err := getLoc(repo.GetFullName())
		if err != nil {
			panic(err)
		}
		fmt.Printf(spin.ClearLine)
		fmt.Printf("%s - %d lines\n", repo.GetFullName(), l.Lines)
		total.merge(l)
	}
	s.Stop()
	fmt.Println(spin.ClearLine)
	fmt.Printf("Total LOC stats:\n"+
		" - %d lines total\n"+
		" - %d lines of code\n"+
		" - %d blank lines\n"+
		" - %d comment lines\n"+
		" - %d files total\n",
		total.Lines, total.Code, total.Blanks,
		total.Comments, total.Files)
}

type loc struct {
	Language string `json:"language"`
	Code     int    `json:"linesOfCode,string"`
	Files    int    `json:"files,string"`
	Blanks   int    `json:"blanks,string"`
	Comments int64  `json:"comments,string"`
	Lines    int64  `json:"lines,string"`
}

func (l *loc) merge(upd *loc) {
	l.Code += upd.Code
	l.Files += upd.Files
	l.Blanks += upd.Blanks
	l.Comments += upd.Comments
	l.Lines += upd.Lines
}

func getLoc(repo string) (*loc, error) {
	// sleep 3 seconds to avoid too-many-requests error
	time.Sleep(3 * time.Second)
	var result *loc
	rsp, err := http.Get(fmt.Sprintf("https://api.codetabs.com/v1/loc?github=%s", repo))
	if err != nil {
		return result, err
	}
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return result, err
	}
	var locs []loc
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&locs); err != nil {
		return result, fmt.Errorf("Failed to decode JSON (%s): %s", string(body), err)
	}
	for _, l := range locs {
		if l.Language == "Total" {
			result = &l
			break
		}
	}
	if result == nil {
		return result, fmt.Errorf("Failed to find total LOC")
	}
	return result, nil
}
