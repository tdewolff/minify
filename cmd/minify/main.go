package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/djherbis/atime"
	humanize "github.com/dustin/go-humanize"
	"github.com/matryer/try"
	"github.com/tdewolff/argp"
	min "github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
)

// Version is the current minify version.
var Version = "built from source"

var extMap = map[string]string{
	"asp":         "text/asp",
	"css":         "text/css",
	"ejs":         "text/x-ejs-template",
	"gohtml":      "text/x-go-template",
	"handlebars":  "text/x-handlebars-template",
	"htm":         "text/html",
	"html":        "text/html",
	"js":          "application/javascript",
	"json":        "application/json",
	"mjs":         "application/javascript",
	"mustache":    "text/x-mustache-template",
	"php":         "application/x-httpd-php",
	"rss":         "application/rss+xml",
	"svg":         "image/svg+xml",
	"tmpl":        "text/x-go-template",
	"webmanifest": "application/manifest+json",
	"xhtml":       "application/xhtml-xml",
	"xml":         "text/xml",
}

var (
	help               bool
	hidden             bool
	list               bool
	m                  *min.M
	matches            []string
	matchesRegexp      []*regexp.Regexp
	filters            []string
	filtersRegexp      []*regexp.Regexp
	extensions         map[string]string
	recursive          bool
	quiet              bool
	verbose            int
	version            bool
	watch              bool
	sync               bool
	bundle             bool
	preserve           []string
	preserveMode       bool
	preserveOwnership  bool
	preserveTimestamps bool
	preserveLinks      bool
	mimetype           string
	oldmimetype        string
)

type Matches struct {
	matches *[]string
}

func (scanner Matches) Scan(s []string) (int, error) {
	n := 0
	for _, item := range s {
		if strings.HasPrefix(item, "-") {
			break
		}
		*scanner.matches = append(*scanner.matches, item)
		n++
	}
	return n, nil
}

func (typenamer Matches) TypeName() string {
	return "[]string"
}

type Includes struct {
	filters *[]string
}

func (scanner Includes) Scan(s []string) (int, error) {
	n := 0
	for _, item := range s {
		if strings.HasPrefix(item, "-") {
			break
		}
		*scanner.filters = append(*scanner.filters, "+"+item)
		n++
	}
	return n, nil
}

func (typenamer Includes) TypeName() string {
	return "[]string"
}

type Excludes struct {
	filters *[]string
}

func (scanner Excludes) Scan(s []string) (int, error) {
	n := 0
	for _, item := range s {
		if strings.HasPrefix(item, "-") {
			break
		}
		*scanner.filters = append(*scanner.filters, "-"+item)
		n++
	}
	return n, nil
}

func (typenamer Excludes) TypeName() string {
	return "[]string"
}

// Task is a minify task.
type Task struct {
	root string
	srcs []string
	dst  string
	sync bool
}

// NewTask returns a new Task.
func NewTask(root, input, output string, sync bool) (Task, error) {
	if len(output) != 0 && (output == "." || output[len(output)-1] == os.PathSeparator) {
		rel, err := filepath.Rel(root, input)
		if err != nil {
			return Task{}, err
		}
		output = filepath.Join(output, rel)
	}
	return Task{root, []string{input}, output, sync}, nil
}

// Loggers.
var (
	Error   *log.Logger
	Warning *log.Logger
	Info    *log.Logger
)

func main() {
	// os.Exit doesn't execute pending defer calls, this is fixed by encapsulating run()
	os.Exit(run())
}

func run() int {
	var inputs []string
	var output string
	var siteurl string

	cssMinifier := css.Minifier{}
	htmlMinifier := html.Minifier{}
	jsMinifier := js.Minifier{}
	jsonMinifier := json.Minifier{}
	svgMinifier := svg.Minifier{}
	xmlMinifier := xml.Minifier{}

	defaultPreserve := []string{"mode", "timestamps"}
	if supportsGetOwnership {
		defaultPreserve = []string{"mode", "ownership", "timestamps"}
	}

	f := argp.New("minify")
	f.AddRest(&inputs, "inputs", "Input files or directories, leave blank to use stdin")
	f.AddOpt(&output, "o", "output", nil, "Output file or directory, leave blank to use stdout")
	f.AddOpt(&oldmimetype, "", "mime", nil, "Mimetype (eg. text/css), optional for input filenames (DEPRECATED, use --type)")
	f.AddOpt(&mimetype, "", "type", nil, "Filetype (eg. css or text/css), optional when specifying inputs")
	f.AddOpt(Matches{&matches}, "", "match", nil, "Filename matching pattern, only matching filenames are processed")
	f.AddOpt(Includes{&filters}, "", "include", nil, "Path inclusion pattern, includes paths previously excluded")
	f.AddOpt(Excludes{&filters}, "", "exclude", nil, "Path exclusion pattern, excludes paths from being processed")
	f.AddOpt(&extensions, "", "ext", nil, "Filename extension mapping to filetype (eg. css or text/css)")
	f.AddOpt(&recursive, "r", "recursive", false, "Recursively minify directories")
	f.AddOpt(&hidden, "a", "all", false, "Minify all files, including hidden files and files in hidden directories")
	f.AddOpt(&list, "l", "list", false, "List all accepted filetypes")
	f.AddOpt(&quiet, "q", "quiet", false, "Quiet mode to suppress all output")
	f.AddOpt(argp.Count{&verbose}, "v", "verbose", nil, "Verbose mode, set twice for more verbosity")
	f.AddOpt(&watch, "w", "watch", false, "Watch files and minify upon changes")
	f.AddOpt(&sync, "s", "sync", false, "Copy all files to destination directory and minify when filetype matches")
	f.AddOpt(&preserve, "p", "preserve", defaultPreserve, "Preserve options (mode, ownership, timestamps, links, all)")
	f.AddOpt(&bundle, "b", "bundle", false, "Bundle files by concatenation into a single file")
	f.AddOpt(&version, "", "version", false, "Version")

	f.AddOpt(&siteurl, "", "url", nil, "URL of file to enable URL minification")
	f.AddOpt(&cssMinifier.Precision, "", "css-precision", 0, "Number of significant digits to preserve in numbers, 0 is all")
	f.AddOpt(&htmlMinifier.KeepComments, "", "html-keep-comments", false, "Preserve all comments")
	f.AddOpt(&htmlMinifier.KeepConditionalComments, "", "html-keep-conditional-comments", false, "Preserve all IE conditional comments (DEPRECATED)")
	f.AddOpt(&htmlMinifier.KeepSpecialComments, "", "html-keep-special-comments", false, "Preserve all IE conditionals and SSI tags")
	f.AddOpt(&htmlMinifier.KeepDefaultAttrVals, "", "html-keep-default-attrvals", false, "Preserve default attribute values")
	f.AddOpt(&htmlMinifier.KeepDocumentTags, "", "html-keep-document-tags", false, "Preserve html, head and body tags")
	f.AddOpt(&htmlMinifier.KeepEndTags, "", "html-keep-end-tags", false, "Preserve all end tags")
	f.AddOpt(&htmlMinifier.KeepWhitespace, "", "html-keep-whitespace", false, "Preserve whitespace characters but still collapse multiple into one")
	f.AddOpt(&htmlMinifier.KeepQuotes, "", "html-keep-quotes", false, "Preserve quotes around attribute values")
	f.AddOpt(&jsMinifier.Precision, "", "js-precision", 0, "Number of significant digits to preserve in numbers, 0 is all")
	f.AddOpt(&jsMinifier.KeepVarNames, "", "js-keep-var-names", false, "Preserve original variable names")
	f.AddOpt(&jsMinifier.Version, "", "js-version", 0, "ECMAScript version to toggle supported optimizations (e.g. 2019, 2020), by default 0 is the latest version")
	f.AddOpt(&jsonMinifier.Precision, "", "json-precision", 0, "Number of significant digits to preserve in numbers, 0 is all")
	f.AddOpt(&jsonMinifier.KeepNumbers, "", "json-keep-numbers", false, "Preserve original numbers instead of minifying them")
	f.AddOpt(&svgMinifier.KeepComments, "", "svg-keep-comments", false, "Preserve all comments")
	f.AddOpt(&svgMinifier.Precision, "", "svg-precision", 0, "Number of significant digits to preserve in numbers, 0 is all")
	f.AddOpt(&xmlMinifier.KeepWhitespace, "", "xml-keep-whitespace", false, "Preserve whitespace characters but still collapse multiple into one")
	f.Parse()

	if version {
		if !quiet {
			fmt.Printf("minify %s\n", Version)
		}
		return 0
	}

	for ext, filetype := range extensions {
		if mimetype, ok := extMap[filetype]; ok {
			filetype = mimetype
		}
		extMap[ext] = filetype
	}

	if list {
		if !quiet {
			n := 0
			var keys []string
			for k := range extMap {
				keys = append(keys, k)
				if n < len(k) {
					n = len(k)
				}
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Println(k + strings.Repeat(" ", n-len(k)+2) + extMap[k])
			}
		}
		return 0
	}

	if len(inputs) == 1 && inputs[0] == "-" {
		inputs = inputs[:0] // stdin
	} else if output == "-" {
		output = "" // stdout
	}
	useStdin := len(inputs) == 0

	Error = log.New(ioutil.Discard, "", 0)
	Warning = log.New(ioutil.Discard, "", 0)
	Info = log.New(ioutil.Discard, "", 0)
	if !quiet {
		Error = log.New(os.Stderr, "ERROR: ", 0)
		if 0 < verbose {
			Warning = log.New(os.Stderr, "WARNING: ", 0)
		}
		if 1 < verbose {
			Info = log.New(os.Stderr, "INFO: ", 0)
		}
	}

	// compile matches and regexps
	var err error
	if 0 < len(matches) {
		matchesRegexp = make([]*regexp.Regexp, len(matches))
		for i, pattern := range matches {
			if matchesRegexp[i], err = compilePattern(pattern); err != nil {
				Error.Println(err)
				return 1
			}
		}
	}
	if 0 < len(filters) {
		filtersRegexp = make([]*regexp.Regexp, len(filters))
		for i, pattern := range filters {
			if filtersRegexp[i], err = compilePattern(pattern[1:]); err != nil {
				Error.Println(err)
				return 1
			}
		}
	}

	// detect mimetype, mimetype=="" means we'll infer mimetype from file extensions
	if oldmimetype != "" {
		Error.Printf("deprecated use of '--mime %v', please use '--type %v' instead", oldmimetype, oldmimetype)
		mimetype = oldmimetype
	}
	if slash := strings.Index(mimetype, "/"); slash == -1 && 0 < len(mimetype) {
		var ok bool
		if mimetype, ok = extMap[mimetype]; !ok {
			Error.Println("unknown filetype", mimetype)
			return 1
		}
	}

	if (useStdin || output == "") && (watch || sync) {
		if watch {
			Error.Println("--watch doesn't work with stdin and stdout, specify input and output")
		}
		if sync {
			Error.Println("--sync doesn't work with stdin and stdout, specify input and output")
		}
		return 1
	} else if useStdin && (bundle || recursive) {
		if bundle {
			Error.Println("--bundle doesn't work with stdin, specify input")
		}
		if recursive {
			Error.Println("--recursive doesn't work with stdin, specify input")
		}
		return 1
	} else if output == "" && recursive && !bundle {
		Error.Println("--recursive doesn't work with stdout, specify output or use --bundle")
		return 1
	}
	if mimetype == "" && useStdin {
		Error.Println("must specify --type for stdin")
		return 1
	} else if mimetype != "" && sync {
		Error.Println("must specify either --sync or --type")
		return 1
	}
	if mimetype == "" {
		if !recursive {
			okAll := true
			for _, input := range inputs {
				ext := filepath.Ext(input)
				if 0 < len(ext) {
					ext = ext[1:]
				}
				if _, ok := extMap[ext]; !ok {
					Error.Println("cannot infer mimetype from extension in", input, ", set --type explicitly")
					okAll = false
				}
			}
			if !okAll {
				return 1
			}
		}
		Info.Println("infer mimetype from file extensions")
	} else {
		Info.Println("use mimetype", mimetype)
	}
	if f.IsSet("preserve") {
		if bundle {
			Error.Println("--preserve cannot be used together with --bundle")
			return 1
		} else if useStdin || output == "" {
			Error.Println("--preserve cannot be used together with stdin or stdout")
			return 1
		}
	}
	for _, option := range preserve {
		switch option {
		case "all":
			preserveMode = true
			preserveOwnership = true
			preserveTimestamps = true
			preserveLinks = true
		case "mode":
			preserveMode = true
		case "ownership":
			preserveOwnership = true
		case "timestamps":
			preserveTimestamps = true
		case "links":
			preserveLinks = true
		}
	}
	if preserveOwnership && !supportsGetOwnership {
		Warning.Println(fmt.Errorf("preserve ownership not supported on platform"))
	}

	////////////////

	for i, input := range inputs {
		if input == "-" {
			Error.Println("cannot mix files and stdin as input")
			return 1
		}
		inputs[i] = filepath.Clean(input)
		if input[len(input)-1] == os.PathSeparator {
			inputs[i] += string(os.PathSeparator)
		}
	}

	// set output file or directory, empty means stdout
	dirDst := false
	if output != "" {
		dirDst = IsDir(output)
		if !dirDst {
			if 1 < len(inputs) && !bundle {
				Error.Printf("stat %v: no such file or directory\n", output)
				return 1
			} else if len(inputs) == 1 {
				if info, err := os.Lstat(inputs[0]); err == nil && !bundle && info.Mode().IsDir() && info.Mode()&os.ModeSymlink == 0 {
					dirDst = true
				}
			}
		}
		if dirDst && bundle {
			Error.Println("--bundle requires destination to be stdout or a file")
			return 1
		}

		output = filepath.Clean(output)
		if dirDst {
			output += string(os.PathSeparator)
		}
	} else if 1 < len(inputs) {
		Error.Println("must specify --bundle for multiple input files with stdout destination")
		return 1
	}
	if output == "" {
		Info.Println("minify to stdout")
	} else if !dirDst {
		Info.Println("minify to output file", output)
	} else if output == "."+string(os.PathSeparator) {
		Info.Println("minify to current working directory")
	} else {
		Info.Println("minify to output directory", output)
	}
	if useStdin {
		Info.Println("minify from stdin")
	}

	var tasks []Task
	var roots []string
	if useStdin {
		task, err := NewTask("", "", output, false)
		if err != nil {
			Error.Println(err)
			return 1
		}
		tasks = append(tasks, task)
		roots = append(roots, "")
	} else {
		fsys := NewFS()
		tasks, roots, err = createTasks(fsys, inputs, output)
		if err != nil {
			Error.Println(err)
			return 1
		}
	}

	// concatenate
	if 1 < len(tasks) && bundle {
		// Task.sync == false because dirDst == false
		for _, task := range tasks[1:] {
			tasks[0].srcs = append(tasks[0].srcs, task.srcs[0])
		}
		tasks = tasks[:1]
	}

	// make output directory
	if dirDst {
		if err := os.MkdirAll(output, 0777); err != nil {
			Error.Println(err)
			return 1
		}
	}

	////////////////

	m = min.New()
	m.Add("text/css", &cssMinifier)
	m.Add("text/html", &htmlMinifier)
	m.Add("image/svg+xml", &svgMinifier)
	m.AddRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma|j|live)script(1\\.[0-5])?$|^module$"), &jsMinifier)
	m.AddRegexp(regexp.MustCompile("[/+]json$"), &jsonMinifier)
	m.AddRegexp(regexp.MustCompile("[/+]xml$"), &xmlMinifier)

	aspMinifier := htmlMinifier
	aspMinifier.TemplateDelims = [2]string{"<%", "%>"}
	m.Add("text/asp", &aspMinifier)
	m.Add("text/x-ejs-template", &aspMinifier)

	phpMinifier := htmlMinifier
	phpMinifier.TemplateDelims = [2]string{"<?", "?>"} // also handles <?php
	m.Add("application/x-httpd-php", &phpMinifier)

	tmplMinifier := htmlMinifier
	tmplMinifier.TemplateDelims = [2]string{"{{", "}}"}
	m.Add("text/x-go-template", &tmplMinifier)
	m.Add("text/x-mustache-template", &tmplMinifier)
	m.Add("text/x-handlebars-template", &tmplMinifier)

	if m.URL, err = url.Parse(siteurl); err != nil {
		Error.Println(err)
		return 1
	}

	fails := 0
	start := time.Now()
	if !watch && (len(tasks) == 1 || 0 < verbose) {
		for _, task := range tasks {
			if ok := minify(task); !ok {
				fails++
			}
		}
	} else {
		numWorkers := runtime.NumCPU()
		if 0 < verbose {
			numWorkers = 1
		} else if numWorkers < 4 {
			numWorkers = 4
		}

		chanTasks := make(chan Task, 20)
		chanFails := make(chan int, numWorkers)
		for n := 0; n < numWorkers; n++ {
			go minifyWorker(chanTasks, chanFails)
		}

		if !watch {
			for _, task := range tasks {
				chanTasks <- task
			}
		} else {
			watcher, err := NewWatcher(recursive)
			if err != nil {
				Error.Println(err)
				return 1
			}
			defer watcher.Close()
			changes := watcher.Run()

			for _, filename := range inputs {
				watcher.AddPath(filename)
			}

			for _, task := range tasks {
				watcher.IgnoreNext(task.dst)
				chanTasks <- task
			}

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			for changes != nil {
				select {
				case <-c:
					watcher.Close()
				case file, ok := <-changes:
					if !ok {
						changes = nil
						break
					}
					file = filepath.Clean(file)

					// find longest common path among roots
					root := ""
					for _, path := range roots {
						pathRel, err1 := filepath.Rel(path, file)
						rootRel, err2 := filepath.Rel(root, file)
						if err2 != nil || err1 == nil && len(pathRel) < len(rootRel) {
							root = path
						}
					}

					task, err := NewTask(root, file, output, !fileMatches(file))
					if err != nil {
						Error.Println(err)
						return 1
					}
					watcher.IgnoreNext(task.dst) // skip change on output
					chanTasks <- task
				}
			}
		}

		close(chanTasks)
		for n := 0; n < numWorkers; n++ {
			fails += <-chanFails
		}
	}

	if !watch {
		Info.Println("finished in", time.Since(start))
	}
	if 0 < fails {
		return 1
	}
	return 0
}

func minifyWorker(chanTasks <-chan Task, chanFails chan<- int) {
	fails := 0
	for task := range chanTasks {
		if ok := minify(task); !ok {
			fails++
		}
	}
	chanFails <- fails
}

// compilePattern returns *regexp.Regexp or glob.Glob
func compilePattern(pattern string) (*regexp.Regexp, error) {
	if len(pattern) == 0 || pattern[0] != '~' {
		if strings.HasPrefix(pattern, `\~`) {
			pattern = pattern[1:]
		}
		pattern = regexp.QuoteMeta(pattern)
		pattern = strings.ReplaceAll(pattern, `\*\*`, `.*`)
		pattern = strings.ReplaceAll(pattern, `\*`, fmt.Sprintf(`[^%c]*`, filepath.Separator))
		pattern = strings.ReplaceAll(pattern, `\?`, fmt.Sprintf(`[^%c]?`, filepath.Separator))
		pattern = "^" + pattern + "$"
	}
	return regexp.Compile(pattern)
}

func fileFilter(filename string) bool {
	if 0 < len(matches) {
		match := false
		base := filepath.Base(filename)
		for _, re := range matchesRegexp {
			if re.MatchString(base) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	match := true
	for i, re := range filtersRegexp {
		if re.MatchString(filename) {
			match = filters[i][0] == '+'
		}
	}
	return match
}

func fileMatches(filename string) bool {
	if !fileFilter(filename) {
		return false
	} else if mimetype != "" {
		return true
	}

	ext := filepath.Ext(filename)
	if 0 < len(ext) {
		ext = ext[1:]
	}
	if _, ok := extMap[ext]; !ok {
		return false
	}
	return true
}

func createTasks(fsys fs.FS, inputs []string, output string) ([]Task, []string, error) {
	tasks := []Task{}
	roots := []string{}
	for _, input := range inputs {
		root := filepath.Clean(filepath.Dir(input))
		input = filepath.Clean(input)

		var err error
		var info os.FileInfo
		if !preserveLinks {
			// follow and dereference symlinks
			info, err = fs.Stat(fsys, input)
		} else {
			info, err = os.Lstat(input)
		}
		if err != nil {
			return nil, nil, err
		}

		if preserveLinks && info.Mode()&os.ModeSymlink != 0 {
			// copy symlink as is
			if !sync {
				Warning.Println("--sync not specified, omitting symbolic link", input)
				continue
			}
			task, err := NewTask(root, input, output, true)
			if err != nil {
				return nil, nil, err
			}
			tasks = append(tasks, task)
		} else if info.Mode().IsRegular() {
			valid := fileFilter(input) // don't filter mimetype
			if valid || sync {
				task, err := NewTask(root, input, output, !valid)
				if err != nil {
					return nil, nil, err
				}
				tasks = append(tasks, task)
			}
		} else if info.Mode().IsDir() {
			if !recursive {
				Warning.Println("--recursive not specified, omitting directory", input)
				continue
			}

			var walkFn func(string, fs.DirEntry, error) error
			walkFn = func(input string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				} else if d.Name() == "." || d.Name() == ".." {
					return nil
				} else if d.Name() == "" || !hidden && d.Name()[0] == '.' {
					if d.IsDir() {
						return fs.SkipDir
					}
					return nil
				}

				if !preserveLinks && d.Type()&os.ModeSymlink != 0 {
					// follow and dereference symlinks
					info, err := fs.Stat(fsys, input)
					if err != nil {
						return err
					}
					if info.IsDir() {
						return fs.WalkDir(fsys, input, walkFn)
					}
					d = fs.FileInfoToDirEntry(info)
				}

				if preserveLinks && d.Type()&os.ModeSymlink != 0 {
					// copy symlink as is
					if !sync {
						Warning.Println("--sync not specified, omitting symbolic link", input)
						return nil
					}
					task, err := NewTask(root, input, output, true)
					if err != nil {
						return err
					}
					tasks = append(tasks, task)
				} else if d.Type().IsRegular() {
					valid := fileMatches(input)
					if valid || sync {
						task, err := NewTask(root, input, output, !valid)
						if err != nil {
							return err
						}
						tasks = append(tasks, task)
					}
				}
				return nil
			}
			if err := fs.WalkDir(fsys, input, walkFn); err != nil {
				return nil, nil, err
			}
			roots = append(roots, root)
		} else {
			return nil, nil, fmt.Errorf("not a file or directory %s", input)
		}
	}
	return tasks, roots, nil
}

func minify(t Task) bool {
	// synchronizing files that are not minified but just copied to the same directory, no action needed
	if t.sync {
		if t.srcs[0] == t.dst {
			return true
		} else if info, err := os.Lstat(t.srcs[0]); preserveLinks && err == nil && info.Mode()&os.ModeSymlink != 0 {
			src, err := os.Readlink(t.srcs[0])
			if err != nil {
				Error.Println(err)
				return false
			}
			if err := createSymlink(src, t.dst); err != nil {
				Error.Println(err)
				return false
			}
			return true
		}
	}

	fileMimetype := mimetype
	if mimetype == "" && !t.sync {
		for _, src := range t.srcs {
			ext := filepath.Ext(src)
			if 0 < len(ext) {
				ext = ext[1:]
			}
			srcMimetype, ok := extMap[ext]
			if !ok {
				Warning.Println("cannot infer mimetype from extension in", src, ", set --type explicitly")
				return false
			}
			if fileMimetype == "" {
				fileMimetype = srcMimetype
			} else if srcMimetype != fileMimetype {
				Warning.Println("inferred mimetype", srcMimetype, "of", src, "for concatenation unequal to previous mimetypes, set --type explicitly")
				return false
			}
		}
	}

	srcName := strings.Join(t.srcs, " + ")
	if len(t.srcs) > 1 {
		srcName = "(" + srcName + ")"
	}
	if srcName == "" {
		srcName = "stdin"
	}
	dstName := t.dst
	if dstName == "" {
		dstName = "stdout"
	} else {
		// rename original when overwriting
		for i := range t.srcs {
			if sameFile, _ := SameFile(t.srcs[i], t.dst); sameFile {
				t.srcs[i] += ".bak"
				err := try.Do(func(attempt int) (bool, error) {
					ferr := os.Rename(t.dst, t.srcs[i])
					return attempt < 5, ferr
				})
				if err != nil {
					Error.Println(err)
					return false
				}
				break
			}
		}
	}

	var err error
	var fr io.ReadCloser
	var fw io.WriteCloser
	if len(t.srcs) == 1 {
		fr, err = openInputFile(t.srcs[0])
	} else {
		var sep []byte
		if err == nil && fileMimetype == extMap["js"] {
			sep = []byte(";\n")
		}
		fr, err = openInputFiles(t.srcs, sep)
	}
	if err != nil {
		Error.Println(err)
		return false
	}

	fw, err = openOutputFile(t.dst)
	if err != nil {
		Error.Println(err)
		fr.Close()
		return false
	}

	// synchronize file
	if t.sync {
		_, err = io.Copy(fw, fr)
		fr.Close()
		fw.Close()
		if err != nil {
			Error.Println(err)
			return false
		}
		preserveAttributes(t.srcs[0], t.root, t.dst)
		Info.Println("copy", srcName, "to", dstName)
		return true
	}

	b, err := ioutil.ReadAll(fr)
	if err != nil {
		fr.Close()
		fw.Close()
		Error.Println("cannot minify "+srcName+":", err)
		return false
	}
	w := bytes.NewBuffer(make([]byte, 0, len(b)))

	success := true
	startTime := time.Now()
	if err = m.Minify(fileMimetype, w, bytes.NewReader(b)); err != nil {
		w = bytes.NewBuffer(b) // copy original
		Error.Println("cannot minify "+srcName+":", err)
		success = false
	}

	rLen, wLen := len(b), w.Len()
	_, err = io.Copy(fw, w)
	fr.Close()
	fw.Close()

	if !quiet {
		dur := time.Since(startTime)
		speed := "Inf MB"
		if 0 < dur {
			speed = humanize.Bytes(uint64(float64(rLen) / dur.Seconds()))
		}
		ratio := 1.0
		if 0 < rLen {
			ratio = float64(wLen) / float64(rLen)
		}

		stats := fmt.Sprintf("(%9v, %6v, %6v, %5.1f%%, %6v/s)", dur, humanize.Bytes(uint64(rLen)), humanize.Bytes(uint64(wLen)), ratio*100, speed)
		if srcName != dstName {
			fmt.Println(stats, "-", srcName, "to", dstName)
		} else {
			fmt.Println(stats, "-", srcName)
		}
	}

	// remove original that was renamed, when overwriting files
	for i := range t.srcs {
		if t.srcs[i] == t.dst+".bak" {
			if err == nil {
				if err = os.Remove(t.srcs[i]); err != nil {
					Error.Println(err)
					return false
				}
			} else {
				if err = os.Remove(t.dst); err != nil {
					Error.Println(err)
					return false
				} else if err = os.Rename(t.srcs[i], t.dst); err != nil {
					Error.Println(err)
					return false
				}
			}
			t.srcs[i] = t.dst
			break
		}
	}
	preserveAttributes(t.srcs[0], t.root, t.dst)
	return success
}

func preserveAttributes(src, root, dst string) {
	if src == "" || dst == "" {
		return
	}

	// make sure we only set attributes on directories and files inside the root destination
	var err error
	src, err = filepath.Rel(root, src)
	if err != nil {
		// should never occur
		Error.Printf("src is not part of root path: src=%s root=%s", src, root)
		return
	}

Next:
	srcInfo, err := os.Stat(filepath.Join(root, src))
	if err != nil {
		Warning.Println(err)
		return
	}

	if preserveMode {
		err = os.Chmod(dst, srcInfo.Mode().Perm())
		if err != nil {
			Warning.Println(err)
		}
	}
	if preserveOwnership {
		if uid, gid, ok := getOwnership(srcInfo); ok {
			err = os.Chown(dst, uid, gid)
			if err != nil {
				Warning.Println(err)
			}
		}
	}
	if preserveTimestamps {
		err = os.Chtimes(dst, atime.Get(srcInfo), srcInfo.ModTime())
		if err != nil {
			Warning.Println(err)
		}
	}

	src = filepath.Dir(src)
	dst = filepath.Dir(dst)
	if src != "." {
		// go up to but excluding the root path
		goto Next
	}
}
