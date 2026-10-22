// Command sync regenerates the generated parts of this repository from the
// releases pinned in versions.env.
//
//	go run ./tools/sync -dsl ../dsl -examples ../examples -synod synod
//	go run ./tools/sync -check ...
//	go run ./tools/sync -blocks DIR
//	go run ./tools/sync -notes v0.2.0
//
// The flags may also be given as the environment variables SYNC_DSL,
// SYNC_EXAMPLES and SYNC_SYNOD. -check writes nothing and exits 1 on any
// difference, on a hand-written file stamped with another dsl tag, and on a
// plugin.json version that is not the newest CHANGELOG version. -blocks writes
// every whole-file synod block of the skill files into a module in DIR, so CI
// can run synod fmt --check and synod check on them. -notes prints the CHANGELOG
// section of a release tag, for the GitHub release text.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	refDir      = "skills/synod/reference"
	skillPath   = "skills/synod/SKILL.md"
	agentsPath  = "AGENTS.md"
	examplesDir = refDir + "/examples"
	dslRepo     = "https://github.com/synodlang/dsl"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "sync:", err)
		os.Exit(1)
	}
}

func envOr(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}

func run(args []string) error {
	fl := flag.NewFlagSet("sync", flag.ContinueOnError)
	root := fl.String("root", ".", "repository root")
	check := fl.Bool("check", false, "write nothing; exit 1 on any difference")
	dsl := fl.String("dsl", envOr("SYNC_DSL", ""), "checkout of dsl at DSL_VERSION")
	examples := fl.String("examples", envOr("SYNC_EXAMPLES", ""), "checkout of examples at EXAMPLES_COMMIT")
	synod := fl.String("synod", envOr("SYNC_SYNOD", "synod"), "the pinned synod command")
	blocks := fl.String("blocks", "", "write the whole-file synod blocks of the skill files as a module into this folder, then stop")
	notesTag := fl.String("notes", "", "print the CHANGELOG section of this release tag, then stop")
	if err := fl.Parse(args); err != nil {
		return err
	}
	if *notesTag != "" {
		return printNotes(*root, *notesTag)
	}
	if *blocks != "" {
		return writeBlocks(*root, *blocks)
	}
	if *dsl == "" || *examples == "" {
		return errors.New("need -dsl and -examples (or SYNC_DSL and SYNC_EXAMPLES)")
	}
	pins, err := readPins(filepath.Join(*root, "versions.env"))
	if err != nil {
		return err
	}
	var problems []string
	problems = append(problems, checkStamps(*root, pins["DSL_VERSION"])...)
	problems = append(problems, checkPlugin(*root)...)

	files, err := generate(*root, pins, *dsl, *examples, *synod)
	if err != nil {
		return err
	}
	if *check {
		problems = append(problems, diff(*root, files)...)
		if len(problems) > 0 {
			for _, p := range problems {
				fmt.Fprintln(os.Stderr, "sync:", p)
			}
			fmt.Fprintf(os.Stderr, "sync: %d problem(s); run go run ./tools/sync to regenerate\n", len(problems))
			os.Exit(1)
		}
		fmt.Printf("sync: %d generated file(s) are up to date\n", len(files))
		return nil
	}
	if len(problems) > 0 {
		// Stamps and versions are for people to fix; regeneration still runs.
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, "sync: warning:", p)
		}
	}
	if err := os.RemoveAll(filepath.Join(*root, filepath.FromSlash(examplesDir))); err != nil {
		return err
	}
	for name, b := range files {
		p := filepath.Join(*root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, b, 0o644); err != nil {
			return err
		}
	}
	fmt.Printf("sync: wrote %d generated file(s)\n", len(files))
	return nil
}

func readPins(p string) (map[string]string, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	pins := map[string]string{}
	for _, l := range strings.Split(string(b), "\n") {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		k, v, ok := strings.Cut(l, "=")
		if !ok {
			return nil, fmt.Errorf("%s: bad line %q", p, l)
		}
		pins[k] = v
	}
	for _, k := range []string{"DSL_VERSION", "SYNOD_VERSION", "SYNOD_LINUX_AMD64_SHA256", "EXAMPLES_VERSION", "EXAMPLES_COMMIT"} {
		if pins[k] == "" {
			return nil, fmt.Errorf("%s: %s is missing", p, k)
		}
	}
	return pins, nil
}

var (
	condensedRe = regexp.MustCompile(`^<!-- condensed from dsl (\S+) spec/\d\d -->$`)
	writtenRe   = regexp.MustCompile(`(?i)written for dsl (v[0-9][0-9A-Za-z.\-]*?)\.?(?: -->|$)`)
)

// checkStamps verifies the dsl tag stamped into the hand-written files.
func checkStamps(root, tag string) []string {
	var out []string
	for _, n := range []string{"grammar.md", "phrases.md", "contexts.md"} {
		rel := refDir + "/" + n
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			out = append(out, err.Error())
			continue
		}
		first, _, _ := strings.Cut(string(b), "\n")
		m := condensedRe.FindStringSubmatch(first)
		switch {
		case m == nil:
			out = append(out, rel+": the first line must be <!-- condensed from dsl <tag> spec/0N -->")
		case m[1] != tag:
			out = append(out, fmt.Sprintf("%s: condensed from dsl %s, but DSL_VERSION is %s; review the file and update its stamp", rel, m[1], tag))
		}
	}
	b, err := os.ReadFile(filepath.Join(root, skillPath))
	if err != nil {
		return append(out, err.Error())
	}
	found := 0
	for _, l := range strings.Split(string(b), "\n") {
		if m := writtenRe.FindStringSubmatch(l); m != nil {
			found++
			if m[1] != tag {
				out = append(out, fmt.Sprintf("%s: written for dsl %s, but DSL_VERSION is %s; review the file and update its stamp", skillPath, m[1], tag))
			}
		}
	}
	if found == 0 {
		out = append(out, skillPath+": no \"written for dsl <tag>\" stamp")
	}
	return out
}

var changelogRe = regexp.MustCompile(`^## v(\d+\.\d+\.\d+)(?:\s|$)`)

// checkPlugin verifies that plugin.json's version is the newest CHANGELOG
// version.
func checkPlugin(root string) []string {
	b, err := os.ReadFile(filepath.Join(root, "CHANGELOG.md"))
	if err != nil {
		return []string{err.Error()}
	}
	newest := ""
	for _, l := range strings.Split(string(b), "\n") {
		if m := changelogRe.FindStringSubmatch(l); m != nil {
			newest = m[1]
			break
		}
	}
	if newest == "" {
		return []string{"CHANGELOG.md: no \"## vX.Y.Z\" heading"}
	}
	pb, err := os.ReadFile(filepath.Join(root, ".claude-plugin", "plugin.json"))
	if err != nil {
		return []string{err.Error()}
	}
	var pj struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(pb, &pj); err != nil {
		return []string{".claude-plugin/plugin.json: " + err.Error()}
	}
	if pj.Version != newest {
		return []string{fmt.Sprintf(".claude-plugin/plugin.json: version %q, but the newest CHANGELOG version is %q", pj.Version, newest)}
	}
	return nil
}

// generate returns every generated file, keyed by slash path under root.
func generate(root string, pins map[string]string, dsl, examples, synod string) (map[string][]byte, error) {
	files := map[string][]byte{}
	tag := pins["DSL_VERSION"]

	for _, g := range []struct{ out, spec string }{
		{"canonical.md", "10-canonical-form.md"},
		{"trace-events.md", "08-trace-events.md"},
	} {
		src, err := os.ReadFile(filepath.Join(dsl, "spec", g.spec))
		if err != nil {
			return nil, err
		}
		body := absLinks(string(src), tag, "spec/"+g.spec)
		head := fmt.Sprintf("<!-- Generated by tools/sync from dsl %s spec/%s. Do not edit. -->\n\n", tag, g.spec)
		files[refDir+"/"+g.out] = []byte(head + body)
	}

	if err := checkSynod(synod, pins["SYNOD_VERSION"]); err != nil {
		return nil, err
	}
	for _, g := range []struct{ out, sub string }{
		{"std-catalog.md", "doc"},
		{"errors.md", "explain"},
	} {
		b, err := runSynod(synod, g.sub, "--catalog")
		if err != nil {
			return nil, err
		}
		files[refDir+"/"+g.out] = b
	}

	if err := checkExamples(examples, pins["EXAMPLES_COMMIT"]); err != nil {
		return nil, err
	}
	if err := fs.WalkDir(os.DirFS(examples), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && p != "." {
				return fs.SkipDir
			}
			return nil
		}
		if p == "synod.mod" || p == "README.md" || strings.HasSuffix(p, ".synod") {
			b, err := os.ReadFile(filepath.Join(examples, filepath.FromSlash(p)))
			if err != nil {
				return err
			}
			files[examplesDir+"/"+p] = b
		}
		return nil
	}); err != nil {
		return nil, err
	}

	sk, err := os.ReadFile(filepath.Join(root, skillPath))
	if err != nil {
		return nil, err
	}
	agents, err := agentsFrom(string(sk))
	if err != nil {
		return nil, err
	}
	files[agentsPath] = []byte(agents)
	return files, nil
}

var linkRe = regexp.MustCompile(`(\]\()([^)\s]+)((?:\s+"[^"]*")?\))`)

// absLinks makes the relative links of a spec file absolute at the dsl tag.
// Links with a scheme, and in-page anchors, stay as they are.
func absLinks(src, tag, file string) string {
	base := path.Dir(file)
	return linkRe.ReplaceAllStringFunc(src, func(m string) string {
		sm := linkRe.FindStringSubmatch(m)
		target := sm[2]
		if strings.HasPrefix(target, "#") || strings.HasPrefix(target, "/") {
			return m
		}
		if u, err := url.Parse(target); err != nil || u.Scheme != "" || u.Host != "" {
			return m
		}
		p, frag, _ := strings.Cut(target, "#")
		abs := dslRepo + "/blob/" + tag + "/" + path.Join(base, p)
		if frag != "" {
			abs += "#" + frag
		}
		return sm[1] + abs + sm[3]
	})
}

// agentsFrom turns SKILL.md into AGENTS.md: no front matter, and the
// reference folder spelled from the repository root.
func agentsFrom(skill string) (string, error) {
	rest, ok := strings.CutPrefix(skill, "---\n")
	if !ok {
		return "", errors.New(skillPath + ": no front matter")
	}
	_, body, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		return "", errors.New(skillPath + ": front matter is not closed")
	}
	body = strings.TrimLeft(body, "\n")
	body = strings.ReplaceAll(body, "`reference/` next to this file", "`skills/synod/reference/`")
	body = regexp.MustCompile("`reference/").ReplaceAllString(body, "`skills/synod/reference/")
	body = regexp.MustCompile(`\(reference/`).ReplaceAllString(body, "(skills/synod/reference/")
	body = strings.ReplaceAll(body, "reference/ next to this file", "skills/synod/reference/")
	// Bare file names in the reference table become full paths.
	rowRe := regexp.MustCompile("(?m)^\\| `((?:grammar|phrases|contexts|canonical|std-catalog|trace-events|errors)\\.md|examples/)` \\|")
	body = rowRe.ReplaceAllString(body, "| `skills/synod/reference/$1` |")
	head := "<!-- Generated by tools/sync from skills/synod/SKILL.md. Do not edit. -->\n\n"
	return head + body, nil
}

func checkSynod(synod, want string) error {
	b, err := runSynod(synod, "version")
	if err != nil {
		return err
	}
	if !bytes.Contains(b, []byte("compiler "+want+" ")) {
		return fmt.Errorf("%s is not synod %s:\n%s", synod, want, b)
	}
	return nil
}

func runSynod(synod string, args ...string) ([]byte, error) {
	tmp, err := os.MkdirTemp("", "sync-synod-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	cmd := exec.Command(synod, args...)
	cmd.Dir = tmp
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s %s: %v\n%s", synod, strings.Join(args, " "), err, errb.String())
	}
	return out.Bytes(), nil
}

// checkExamples verifies the checkout when it is a git repository.
func checkExamples(dir, commit string) error {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		return nil
	}
	cmd := exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	b, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git rev-parse in %s: %v", dir, err)
	}
	if got := strings.TrimSpace(string(b)); got != commit {
		return fmt.Errorf("%s is at %s, but EXAMPLES_COMMIT is %s", dir, got, commit)
	}
	return nil
}

// diff compares the generated files with the ones on disk, including files
// under examples/ that should not exist.
func diff(root string, files map[string][]byte) []string {
	var out []string
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(n)))
		switch {
		case errors.Is(err, fs.ErrNotExist):
			out = append(out, n+": missing")
		case err != nil:
			out = append(out, err.Error())
		case !bytes.Equal(got, files[n]):
			out = append(out, n+": differs from the regenerated file")
		}
	}
	filepath.WalkDir(filepath.Join(root, filepath.FromSlash(examplesDir)), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		rel = filepath.ToSlash(rel)
		if _, ok := files[rel]; !ok {
			out = append(out, rel+": not in the pinned examples")
		}
		return nil
	})
	return out
}

var fenceRe = regexp.MustCompile("(?s)(?m)^```synod\n(.*?)^```$")
var pkgRe = regexp.MustCompile(`^# package (\w+)\n`)

// writeBlocks writes every whole-file block (one that starts with
// "# package") of the skill files into a module github.com/acme/demo.
func writeBlocks(root, dir string) error {
	var skills []string
	skills = append(skills, skillPath)
	ents, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(refDir)))
	if err != nil {
		return err
	}
	for _, e := range ents {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			skills = append(skills, refDir+"/"+e.Name())
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "synod.mod"), []byte("module github.com/acme/demo\nsynod 0.2\n"), 0o644); err != nil {
		return err
	}
	seen := map[string]string{}
	n := 0
	for _, s := range skills {
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(s)))
		if err != nil {
			return err
		}
		for _, m := range fenceRe.FindAllStringSubmatch(string(b), -1) {
			pm := pkgRe.FindStringSubmatch(m[1])
			if pm == nil {
				continue
			}
			name := pm[1]
			if prev, ok := seen[name]; ok {
				return fmt.Errorf("%s: package %s is also a block in %s; give the blocks distinct package names", s, name, prev)
			}
			seen[name] = s
			d := filepath.Join(dir, name)
			if err := os.MkdirAll(d, 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(d, name+".synod"), []byte(m[1]), 0o644); err != nil {
				return err
			}
			n++
		}
	}
	fmt.Printf("sync: wrote %d block(s) to %s\n", n, dir)
	return nil
}
