package server

import (
	"context"

	"github.com/seniorGolang/i2s/pkg/tags"
)

const (
	mainTagsContextKey = "MainTags"
	ael                = "AllowEllipsis"
	spi                = "SourcePackageImport"
)

func prepareContext(info *GenerationInfo) context.Context {

	ctx := context.Background()

	ctx = WithTags(ctx, tags.ParseTags(info.Iface.Docs))
	ctx = WithSourcePackageImport(ctx, info.SourcePackageImport)
	return ctx
}

func WithSourcePackageImport(parent context.Context, val string) context.Context {
	return context.WithValue(parent, spi, val)
}

func SourcePackageImport(ctx context.Context) string {
	return ctx.Value(spi).(string)
}

func WithTags(parent context.Context, tt tags.DocTags) context.Context {
	return context.WithValue(parent, mainTagsContextKey, tt)
}

func Tags(ctx context.Context) tags.DocTags {
	return ctx.Value(mainTagsContextKey).(tags.DocTags)
}

func AllowEllipsis(ctx context.Context) bool {
	v, ok := ctx.Value(ael).(bool)
	return ok && v
}
