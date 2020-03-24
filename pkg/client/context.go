package client

import (
	"context"

	"github.com/vetcher/go-astra"
	"github.com/vetcher/go-astra/types"

	"github.com/seniorGolang/i2s/pkg/tags"
)

const (
	mainTagsContextKey = "MainTags"
	ael                = "AllowEllipsis"
	spi                = "SourcePackageImport"
)

func prepareContext(filename string, iface *types.Interface) (context.Context, error) {

	ctx := context.Background()
	p, err := astra.ResolvePackagePath(filename)

	if err != nil {
		return nil, err
	}
	ctx = WithSourcePackageImport(ctx, p)

	set := TagsSet{}
	genTags := tags.ParseTags(iface.Docs)

	for _, tag := range genTags {
		set.Add(tag)
	}

	ctx = WithTags(ctx, set)
	return ctx, nil
}

func WithSourcePackageImport(parent context.Context, val string) context.Context {
	return context.WithValue(parent, spi, val)
}

func SourcePackageImport(ctx context.Context) string {
	return ctx.Value(spi).(string)
}

func WithTags(parent context.Context, tt TagsSet) context.Context {
	return context.WithValue(parent, mainTagsContextKey, tt)
}

func Tags(ctx context.Context) TagsSet {
	return ctx.Value(mainTagsContextKey).(TagsSet)
}

type TagsSet map[string]struct{}

func (s TagsSet) Has(item string) bool {
	_, ok := s[item]
	return ok
}

func (s TagsSet) HasAny(items ...string) bool {
	if len(items) == 0 {
		return false
	}
	return s.Has(items[0]) || s.HasAny(items[1:]...)
}

func (s TagsSet) Add(item string) {
	s[item] = struct{}{}
}

func AllowEllipsis(ctx context.Context) bool {
	v, ok := ctx.Value(ael).(bool)
	return ok && v
}
