package meta

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vetcher/go-astra/types"

	"github.com/seniorGolang/i2s/pkg/tags"
	"github.com/seniorGolang/i2s/pkg/utils"
)

type GenerationInfo struct {
	PkgName             string
	ServiceName         string
	SourceFilePath      string
	OutputFilePath      string
	BasePackageImport   string
	OutputPackageImport string
	SourcePackageImport string

	Iface *types.Interface

	Backend  string
	Services []types.Interface
}

func (i GenerationInfo) String() string {

	var ss []string

	ss = append(ss,
		fmt.Sprint(),
		fmt.Sprint("ServiceName: ", i.ServiceName),
		fmt.Sprint("SourcePackageImport: ", i.SourcePackageImport),
		fmt.Sprint("SourceFilePath: ", i.SourceFilePath),
		fmt.Sprint("OutputPackageImport: ", i.OutputPackageImport),
		fmt.Sprint("OutputFilePath: ", i.OutputFilePath),
		fmt.Sprint(),
	)
	return strings.Join(ss, "\n\t")
}

func listKeysOfMap(m map[string]bool) string {

	var keys = make([]string, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}

func FillInfo(iface *types.Interface, sourcePath, outPath string) (info *GenerationInfo, err error) {

	var absOutPath string
	if absOutPath, err = filepath.Abs(outPath); err != nil {
		return
	}

	var importPackagePath string

	if importPackagePath, err = utils.GetPkgPath(".", true); err != nil {
		return
	}

	// if importPackagePath, err = resolvePackagePath(filepath.Dir(sourcePath)); err != nil {
	//	return
	// }

	var absSourcePath string
	if absSourcePath, err = filepath.Abs(sourcePath); err != nil {
		return
	}

	m := make(map[string]bool, len(iface.Methods))

	for _, fn := range iface.Methods {

		if tags.ParseTags(fn.Docs).IsSet("disable") {
			continue
		}
		m[fn.Name] = true
	}

	pkgImport := path.Dir(path.Join(importPackagePath, sourcePath))
	pkgBase, _ := path.Split(pkgImport + ".file")

	info = &GenerationInfo{
		Iface:               iface,
		OutputFilePath:      absOutPath,
		SourceFilePath:      absSourcePath,
		OutputPackageImport: importPackagePath,
		BasePackageImport:   pkgBase,
		SourcePackageImport: pkgImport,
		PkgName:             strings.ToLower(iface.Name),
		ServiceName:         utils.ToLowerCamel(iface.Name),
	}
	return
}

func resolvePackagePath(outPath string) (path string, err error) {

	goPath := os.Getenv("GOPATH")

	if goPath == "" {
		err = fmt.Errorf("GOPATH is empty")
		return
	}

	var absOutPath string

	if absOutPath, err = filepath.Abs(outPath); err != nil {
		return
	}

	for _, path := range strings.Split(goPath, ":") {

		goPathSrc := filepath.Join(path, "src")

		if strings.HasPrefix(absOutPath, goPathSrc) {
			return absOutPath[len(goPathSrc)+1:], nil
		}
	}
	return
}
