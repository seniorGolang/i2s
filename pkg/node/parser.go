package node

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/vetcher/go-astra"
	"github.com/vetcher/go-astra/types"

	"github.com/seniorGolang/i2s/pkg/tags"
)

func Parse(servicesPath string, ifaceNames ...string) (node Node, err error) {

	var files []os.FileInfo
	if files, err = ioutil.ReadDir(servicesPath); err != nil {
		return
	}

	for _, file := range files {

		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		var services []Service
		var fileTags tags.DocTags

		if services, fileTags, err = ParseGoFile(servicesPath, file.Name(), ifaceNames...); err != nil {
			return
		}

		node.Tags = node.Tags.Merge(fileTags)
		node.Services = append(node.Services, services...)
	}

	node.Name = node.Tags.Value("backend")
	node.Events, _ = ParseEvents(path.Join(servicesPath, "events"))
	return
}

func ParseService(servicesPath string, ifaceNames ...string) (service []Service, events []Event, serviceTags tags.DocTags, err error) {

	var files []os.FileInfo
	if files, err = ioutil.ReadDir(servicesPath); err != nil {
		return
	}

	serviceTags = make(tags.DocTags)

	for _, file := range files {

		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		var services []Service
		var fileTags tags.DocTags

		if services, fileTags, err = ParseGoFile(servicesPath, file.Name(), ifaceNames...); err != nil {
			return
		}

		service = append(service, services...)
		serviceTags = serviceTags.Merge(fileTags)
	}

	events, _ = ParseEvents(path.Join(servicesPath, "events"))
	return
}

func ParseGoFile(servicesPath, fileName string, ifaceNames ...string) (services []Service, fileTags tags.DocTags, err error) {

	var goFile *types.File
	if goFile, err = astra.ParseFile(path.Join(servicesPath, fileName)); err != nil {
		return
	}

	fileTags = tags.ParseTags(goFile.Docs)

	for _, iface := range goFile.Interfaces {

		if len(ifaceNames) == 0 || in(iface.Name, ifaceNames) {

			var svc Service
			if svc, err = parseIface(servicesPath, iface); err != nil {
				return
			}
			services = append(services, svc)
		}
	}
	return
}

func parseIface(pkgPath string, iface types.Interface) (svc Service, err error) {

	svc.Name = iface.Name
	svc.Tags = tags.ParseTags(iface.Docs)

	for _, ifaceMethod := range iface.Methods {

		log.Infof("method %s", ifaceMethod.Name)

		var m Method
		if m, err = parseMethod(pkgPath, ifaceMethod); err != nil {
			return
		}
		svc.Methods = append(svc.Methods, m)
	}
	return
}

func parseMethod(pkgPath string, method *types.Function) (m Method, err error) {

	m.Name = method.Name
	m.Tags = tags.ParseTags(method.Docs)

	vars := func(vars []types.Variable) (objList []Object, err error) {

		for _, v := range vars {

			var obj Object
			if obj, err = parseObject(pkgPath, v); err != nil {
				return
			}
			objList = append(objList, obj)
		}
		return
	}

	var arguments []types.Variable
	arguments, m.HasContext = removeContextIfFirst(method.Args)

	if m.Arguments, err = vars(arguments); err != nil {
		return
	}

	var results []types.Variable
	results, m.ReturnError = removeErrorIfLast(method.Results)

	m.Results, err = vars(results)
	return
}

func in(value string, values []string) bool {

	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
