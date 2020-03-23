package pkgpath

import (
	"bytes"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/seniorGolang/i2s/pkg/logger"
)

var (
	log = logger.Log.WithField("module", "server")
)

func GetPkgPath(fName string, isDir bool) (string, error) {

	// find go.mod file
	goModPath, err := GoModPath(fName, isDir)
	if err != nil {
		log.Error(errors.Wrap(err, "cannot find go.mod because of"))
	}

	if strings.Contains(goModPath, "go.mod") {
		pkgPath, err := GetPkgPathFromGoMod(fName, isDir, goModPath)
		if err != nil {
			return "", err
		}
		return pkgPath, nil
	}
	return GetPkgPathFromGOPATH(fName, isDir)
}

var (
	goModPathCache = make(map[string]string)
)

func GoModPath(fName string, isDir bool) (string, error) {

	root := fName

	if !isDir {
		root = filepath.Dir(fName)
	}

	goModPath, ok := goModPathCache[root]

	if ok {
		return goModPath, nil
	}

	defer func() {
		goModPathCache[root] = goModPath
	}()

	var stdout []byte
	var err error

	for {
		cmd := exec.Command("go", "env", "GOMOD")
		cmd.Dir = root
		stdout, err = cmd.Output()

		if err == nil {
			break
		}

		if _, ok := err.(*os.PathError); ok {
			// try to find go.mod on level higher
			r := filepath.Join(root, "..")
			if r == root { // when we in root directory stop trying
				return "", err
			}
			root = r
			continue
		}
		return "", err
	}
	goModPath = string(bytes.TrimSpace(stdout))
	return goModPath, nil
}

func GetPkgPathFromGoMod(fName string, isDir bool, goModPath string) (string, error) {

	modulePath := GetModulePath(goModPath)

	if modulePath == "" {
		return "", errors.Errorf("cannot determine module path from %s", goModPath)
	}

	rel := path.Join(modulePath, filePathToPackagePath(strings.TrimPrefix(fName, filepath.Dir(goModPath))))

	if !isDir {
		return path.Dir(rel), nil
	}
	return path.Clean(rel), nil
}

var (
	gopathCache           = ""
	modulePrefix          = []byte("\nmodule ")
	pkgPathFromGoModCache = make(map[string]string)
)

func GetModulePath(goModPath string) string {

	pkgPath, ok := pkgPathFromGoModCache[goModPath]

	if ok {
		return pkgPath
	}

	defer func() {
		pkgPathFromGoModCache[goModPath] = pkgPath
	}()

	data, err := ioutil.ReadFile(goModPath)

	if err != nil {
		return ""
	}

	var i int

	if bytes.HasPrefix(data, modulePrefix[1:]) {
		i = 0
	} else {
		i = bytes.Index(data, modulePrefix)
		if i < 0 {
			return ""
		}
		i++
	}

	line := data[i:]

	// Cut line at \n, drop trailing \r if present.
	if j := bytes.IndexByte(line, '\n'); j >= 0 {
		line = line[:j]
	}

	if line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}

	line = line[len("module "):]

	// If quoted, unquote.
	pkgPath = strings.TrimSpace(string(line))

	if pkgPath != "" && pkgPath[0] == '"' {
		s, err := strconv.Unquote(pkgPath)
		if err != nil {
			return ""
		}
		pkgPath = s
	}
	return pkgPath
}

func GetRelatedFilePath(pkg string) (string, error) {

	if gopathCache == "" {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			var err error
			gopath, err = GetDefaultGoPath()
			if err != nil {
				return "", errors.Wrap(err, "cannot determine GOPATH")
			}
		}
		gopathCache = gopath
	}

	paths := allPaths(filepath.SplitList(gopathCache))

	for _, p := range paths {
		checkingPath := filepath.Join(p, pkg)
		if info, err := os.Stat(checkingPath); err == nil && info.IsDir() {
			return checkingPath, nil
		}
	}
	return "", errors.Errorf("file '%v' is not in GOROOT or GOPATH. Checked paths:\n%s", pkg, strings.Join(paths, "\n"))
}

func allPaths(gopaths []string) []string {

	const _2 = 2
	res := make([]string, len(gopaths)+_2)
	res[0] = filepath.Join(build.Default.GOROOT, "src")
	res[1] = "vendor"
	for i := range res[_2:] {
		res[i+_2] = filepath.Join(gopaths[i], "src")
	}
	return res
}

func GetPkgPathFromGOPATH(fName string, isDir bool) (string, error) {

	if gopathCache == "" {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			var err error
			gopath, err = GetDefaultGoPath()
			if err != nil {
				return "", errors.Wrap(err, "cannot determine GOPATH")
			}
		}
		gopathCache = gopath
	}

	for _, p := range filepath.SplitList(gopathCache) {
		prefix := filepath.Join(p, "src") + string(filepath.Separator)
		if rel := strings.TrimPrefix(fName, prefix); rel != fName {
			if !isDir {
				return path.Dir(filePathToPackagePath(rel)), nil
			} else {
				return path.Clean(filePathToPackagePath(rel)), nil
			}
		}
	}

	return "", errors.Errorf("file '%s' is not in GOPATH. Checked paths:\n%s", fName, strings.Join(filepath.SplitList(gopathCache), "\n"))
}

func filePathToPackagePath(path string) string {
	return filepath.ToSlash(path)
}

func GetDefaultGoPath() (string, error) {

	if build.Default.GOPATH != "" {
		return build.Default.GOPATH, nil
	}

	output, err := exec.Command("go", "env", "GOPATH").Output()
	return string(bytes.TrimSpace(output)), err
}

func PackageName(path string, decl string) (string, error) {

	pkgs, err := parser.ParseDir(token.NewFileSet(), path, nonTestFilter, parser.PackageClauseOnly)

	if err != nil {
		if os.IsNotExist(err) {
			return filepath.Base(path), nil
		}
		return "", err
	}

	var alias string

	for k, pkg := range pkgs {

		log.Debug(path, "has package", k)

		// Name of type was not provided
		if decl == "" {
			alias = k
			break
		}

		if !ast.PackageExports(pkg) {
			continue
		}

		if ast.FilterPackage(pkg, func(name string) bool { return name == decl }) {
			// filter returns true if package has declaration
			// make it to be sure, that we choose right alias
			alias = k
			break
		}
	}
	return alias, nil
}

// filters all files with tests
func nonTestFilter(info os.FileInfo) bool {
	return !strings.HasSuffix(info.Name(), "_test.go")
}

func GetPkgPrefix(base, pkg string) (prefix string) {

	base, _ = filepath.Abs(base)

	pkgItems := strings.Split(pkg, "/")
	baseItems := strings.Split(base, "/")

	for index, item := range baseItems {

		if pkgItems[index] != item {
			break
		}
	}

	return
}
