package node

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/vetcher/go-astra"
	"github.com/vetcher/go-astra/types"

	"github.com/seniorGolang/i2s/pkg/node/mod"
	"github.com/seniorGolang/i2s/pkg/tags"
)

func parseObject(pkgPath string, v types.Variable) (obj Object, err error) {
	return makeType(pkgPath, v, v.Type)
}

func objectFromStruct(pkgPath string, structInfo types.Struct) (obj Object) {

	obj.Name = structInfo.Name
	obj.Type = structInfo.Base.Name
	obj.Tags = tags.ParseTags(structInfo.Docs)

	for _, fieldInfo := range structInfo.Fields {

		field, err := makeType(pkgPath, fieldInfo.Variable, fieldInfo.Type)

		if err != nil {
			log.Error(fieldInfo, err)
		}

		field.TypeTags = fieldInfo.Tags

		if len(fieldInfo.Name) > 0 {
			field.IsPrivate = string([]rune(fieldInfo.Name)[0]) != strings.ToUpper(string([]rune(fieldInfo.Name)[0]))
		}
		obj.Fields = append(obj.Fields, field)
	}
	return
}

func makeType(pkgPath string, field types.Variable, fieldType types.Type) (obj Object, err error) {

	for fieldType != nil {

		switch f := fieldType.(type) {

		case types.TName:

			if IsBuiltin(fieldType) {
				obj = Object{Name: field.Name, Type: fieldType.String(), Tags: tags.ParseTags(field.Docs)}
				return
			}

			obj, err = searchTypeInfo(pkgPath, f.TypeName)
			obj.Name = field.Name
			return

		case types.Struct:

			if IsBuiltin(fieldType) {
				obj = Object{Name: field.Name, Type: fieldType.String(), Tags: tags.ParseTags(field.Docs)}
				return
			}
			obj, err = searchTypeInfo(pkgPath, f.Name)
			obj.Name = field.Name
			return

		case types.TImport:

			if IsBuiltin(fieldType) {
				obj = Object{Name: field.Name, Type: fieldType.String(), Tags: tags.ParseTags(field.Docs)}
				return
			}
			obj, err = searchTypeInfo(f.Import.Package, f.Next.String())
			obj.Name = field.Name
			return

		case types.TArray:
			obj, err = makeType(pkgPath, field, f.Next)
			obj.IsArray = true
			return

		case types.TEllipsis:
			obj, err = makeType(pkgPath, field, f.Next)
			obj.IsEllipsis = true
			return

		case types.TMap:

			m := fieldType.(types.TMap)

			key, _ := makeType("", field, m.Key)
			val, _ := makeType("", field, m.Value)

			obj = Object{Name: field.Name, IsMap: true, Tags: tags.ParseTags(field.Docs), Type: fmt.Sprintf("map[%s]%s", m.Key, m.Value), SubTypes: map[string]Object{
				"key":   key,
				"value": val,
			}}
			return

		case types.TPointer:
			obj, err = makeType(pkgPath, field, f.Next)
			obj.IsNullable = true
			return

		case types.TInterface:

			obj = Object{Name: field.Name, Tags: tags.ParseTags(field.Docs), Type: "Interface", IsNullable: true}
			return

		default:
			err = errors.New("unknown type " + fieldType.String())
			return
		}
	}
	return
}

func searchTypeInfo(pkg, name string) (obj Object, err error) {

	if obj, err = getStructInfo(pkg, name); err != nil {

		pkgPath := mod.PkgModPath(pkg)

		if obj, err = getStructInfo(pkgPath, name); err != nil {

			pkgPath = path.Join("./vendor", pkg)

			if obj, err = getStructInfo(pkgPath, name); err != nil {

				pkgPath = trimLocalPkg(pkg)
				obj, err = getStructInfo(pkgPath, name)
			}
		}
	}
	return
}

func getStructInfo(relPath, name string) (obj Object, err error) {

	pkgPath, _ := filepath.Abs(relPath)

	err = filepath.Walk(pkgPath, func(filePath string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		var srcFile *types.File
		if srcFile, err = astra.ParseFile(path.Join(pkgPath, info.Name())); err != nil {
			return errors.Wrap(err, fmt.Sprintf("%s,%s", relPath, name))
		}

		for _, structInfo := range srcFile.Structures {

			if structInfo.Name == name {
				obj = objectFromStruct(relPath, structInfo)
				break
			}
		}
		return nil
	})
	return
}

func jsonName(fieldInfo types.StructField) (value string) {

	value = fieldInfo.Name
	if tagValues, _ := fieldInfo.Tags["json"]; len(tagValues) > 0 {
		value = tagValues[0]
	}
	return
}

func getModName() (module string) {

	modFile, err := os.OpenFile("go.mod", os.O_RDONLY, os.ModePerm)

	if err != nil {
		return
	}
	defer modFile.Close()

	rd := bufio.NewReader(modFile)
	if module, err = rd.ReadString('\n'); err != nil {
		return ""
	}
	module = strings.Trim(module, "\n")

	moduleTokens := strings.Split(module, " ")

	if len(moduleTokens) == 2 {
		module = strings.TrimSpace(moduleTokens[1])
	}
	return
}

func trimLocalPkg(pkg string) (pgkPath string) {

	module := getModName()

	if module == "" {
		return pkg
	}

	moduleTokens := strings.Split(module, "/")
	pkgTokens := strings.Split(pkg, "/")

	if len(pkgTokens) < len(moduleTokens) {
		return pkg
	}

	pgkPath = path.Join(strings.Join(pkgTokens[len(moduleTokens):], "/"))
	return
}

func IsBuiltin(t types.Type) bool {

	if types.IsBuiltin(t) {
		return true
	}

	typeString := strings.TrimPrefix(t.String(), "*")

	switch typeString {
	case "uuid.UUID":
		return true
	case "UUID":
		return true
	case "json.RawMessage":
		return true
	case "bson.ObjectId":
		return true
	case "time.Time":
		return true
	}
	return false
}
