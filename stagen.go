/* MIT/X Consortium License

 (c) 2025 sebastian <sm@secpaste.dev>

 Permission is hereby granted, free of charge, to any person obtaining a
 copy of this software and associated documentation files (the "Software"),
 to deal in the Software without restriction, including without limitation
 the rights to use, copy, modify, merge, publish, distribute, sublicense,
 and/or sell copies of the Software, and to permit persons to whom the
 Software is furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included in
 all copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL
 THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
 FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
 DEALINGS IN THE SOFTWARE. */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type Args struct {
	repo    string
	out     string
	name    string
	desc    string
	url     string
}

type Repo struct {
	Name          string
	Desc          string
	Url           string
	LastCommit    string
	Files         []File
	Commits       []Commit
	Refs          []Ref
	Title         string
	StylePath     string
	BasePath      string
	ReadmeContent string
	Dir           string
}

type MainIndex struct {
	Repos []Repo
}

type File struct {
	FileName string
	Path     string
	Size     string
	Mode     string
}

type Commit struct {
	Hash      string
	ShortHash string
	Author    string
	Date      string
	Subject   string
	Body      string
	Files     []string
	Stats     string
}

type Ref struct {
	Name string
	Hash string
	Type string
}

func die(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func run(cmd string, args ...string) string {
	out, err := exec.Command(cmd, args...).Output()
	die(err)
	return strings.TrimSpace(string(out))
}

func runInDir(dir, cmd string, args ...string) string {
	c := exec.Command(cmd, args...)
	c.Dir = dir
	out, err := c.Output()
	die(err)
	return strings.TrimSpace(string(out))
}

func getCommits(repo string) []Commit {
	lines := strings.Split(runInDir(repo, "git", "log", "--format=%H|%h|%an|%ad|%s|%b", "--date=short", "-n", "50"), "\n")
	commits := make([]Commit, 0)
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 5 {
			continue
		}
		
		hash := parts[0]
		stats := runInDir(repo, "git", "show", "--stat", "--format=", hash)
		files := strings.Split(runInDir(repo, "git", "show", "--name-only", "--format=", hash), "\n")
		
		commit := Commit{
			Hash:      hash,
			ShortHash: parts[1],
			Author:    parts[2],
			Date:      parts[3],
			Subject:   parts[4],
			Body:      strings.Join(parts[5:], "|"),
			Files:     files,
			Stats:     stats,
		}
		commits = append(commits, commit)
	}
	return commits
}

func getFiles(repo string) []File {
	lines := strings.Split(runInDir(repo, "git", "ls-tree", "-r", "-l", "HEAD"), "\n")
	files := make([]File, 0)
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue
		}
		
		size := parts[3]
		if parts[1] == "tree" {
			size = "-"
		}
		
		file := File{
			Mode:     parts[0],
			FileName: filepath.Base(parts[4]),
			Path:     parts[4],
			Size:     size,
		}
		files = append(files, file)
	}
	
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
	
	return files
}

func getRefs(repo string) []Ref {
	lines := strings.Split(runInDir(repo, "git", "for-each-ref", "--format=%(refname:short)|%(objectname)|%(objecttype)"), "\n")
	refs := make([]Ref, 0)
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}
		
		ref := Ref{
			Name: parts[0],
			Hash: parts[1],
			Type: parts[2],
		}
		refs = append(refs, ref)
	}
	return refs
}

func getLastCommit(repo string) string {
	return runInDir(repo, "git", "log", "-1", "--format=%ad", "--date=short")
}

func getReadme(repo string) string {
	readmeFiles := []string{"README.md", "README.txt", "README", "readme.md", "readme.txt", "readme"}
	for _, filename := range readmeFiles {
		c := exec.Command("git", "show", "HEAD:"+filename)
		c.Dir = repo
		out, err := c.Output()
		if err == nil && len(out) > 0 {
			return strings.TrimSpace(string(out))
		}
	}
	return ""
}

func mkRepo(args Args) Repo {
	return Repo{
		Name:          args.name,
		Desc:          args.desc,
		Url:           args.url,
		LastCommit:    getLastCommit(args.repo),
		Files:         getFiles(args.repo),
		Commits:       getCommits(args.repo),
		Refs:          getRefs(args.repo),
		Title:         "",
		StylePath:     "/style.css",
		BasePath:      "",
		ReadmeContent: getReadme(args.repo),
		Dir:           filepath.Base(args.out),
	}
}

func writeFile(path string, tmplFile string, data interface{}) {
	os.MkdirAll(filepath.Dir(path), 0755)
	
	t := template.Must(template.ParseGlob("*.tmpl"))
	
	f, err := os.Create(path)
	die(err)
	defer f.Close()
	
	die(t.ExecuteTemplate(f, tmplFile, data))
}

func genIndex(args Args, repo Repo) {
	repo.Title = "Files"
	writeFile(args.out+"/index.html", "index.tmpl", repo)
}

func genLog(args Args, repo Repo) {
	repo.Title = "Log"
	writeFile(args.out+"/log.html", "log.tmpl", repo)
}

func genCommits(args Args, repo Repo) {
	repo.Title = "Commits"
	writeFile(args.out+"/commits.html", "commits.tmpl", repo)
}

func genRefs(args Args, repo Repo) {
	repo.Title = "Refs"
	writeFile(args.out+"/refs.html", "refs.tmpl", repo)
}

func genReadme(args Args, repo Repo) {
	repo.Title = "README"
	writeFile(args.out+"/readme.html", "readme.tmpl", repo)
}

func genFilePages(args Args, repo Repo) {
	for _, file := range repo.Files {
		content := runInDir(args.repo, "git", "show", "HEAD:"+file.Path)
		
		depth := strings.Count(file.Path, "/") + 1
		basePath := strings.Repeat("../", depth)
		
		data := struct {
			Repo
			File
			Content string
		}{repo, file, content}
		data.Title = file.Path
		data.StylePath = "/style.css"
		data.BasePath = basePath
		
		writeFile(args.out+"/file/"+file.Path+".html", "file.tmpl", data)
	}
}

func hlDiff(diff string) string {
	lines := strings.Split(diff, "\n")
	for i, line := range lines {
		if len(line) == 0 {
			continue
		}
		escaped := html.EscapeString(line)
		switch line[0] {
		case '+':
			lines[i] = `<span class="i">` + escaped + `</span>`
		case '-':
			lines[i] = `<span class="d">` + escaped + `</span>`
		default:
			lines[i] = escaped
		}
	}
	return strings.Join(lines, "\n")
}

func genCommitPages(args Args, repo Repo) {
	for _, commit := range repo.Commits {
		diff := runInDir(args.repo, "git", "show", commit.Hash)
		
		data := struct {
			Repo
			Commit
			Diff template.HTML
		}{repo, commit, template.HTML(hlDiff(diff))}
		data.Title = commit.ShortHash
		data.StylePath = "/style.css"
		data.BasePath = "../"
		
		writeFile(args.out+"/commit/"+commit.Hash+".html", "commit.tmpl", data)
	}
}

func updateMainIndex(args Args, repo Repo) {
	parentDir := filepath.Dir(args.out)
	indexPath := parentDir + "/index.json"
	
	var repos []Repo
	data, err := ioutil.ReadFile(indexPath)
	if err == nil {
		json.Unmarshal(data, &repos)
	}
	
	for i, r := range repos {
		if r.Dir == repo.Dir {
			repos[i] = repo
			goto write
		}
	}
	repos = append(repos, repo)
	
write:
	data, _ = json.Marshal(repos)
	ioutil.WriteFile(indexPath, data, 0644)
	
	t := template.Must(template.ParseFiles("main-index.tmpl"))
	f, err := os.Create(parentDir + "/index.html")
	if err != nil {
		return
	}
	defer f.Close()
	
	os.Rename("style.css", parentDir+"/style.css")
	t.Execute(f, MainIndex{repos})
}

func main() {
	repo := flag.String("repo", "", "git repository path (required)")
	out := flag.String("out", "", "output directory (required)")
	name := flag.String("name", "", "repository name (required)")
	desc := flag.String("desc", "", "repository description")
	url := flag.String("url", "", "repository url")
	flag.Parse()
	
	if *repo == "" || *out == "" || *name == "" {
		log.Fatal("repo, out and name are required")
	}
	
	args := Args{*repo, *out, *name, *desc, *url}
	r := mkRepo(args)
	
	os.RemoveAll(*out)
	os.MkdirAll(*out, 0755)
	
	genIndex(args, r)
	genLog(args, r)
	genCommits(args, r)
	genRefs(args, r)
	genReadme(args, r)
	genFilePages(args, r)
	genCommitPages(args, r)
	
	updateMainIndex(args, r)
	
	fmt.Printf("Generated static git viewer in %s\n", *out)
}
