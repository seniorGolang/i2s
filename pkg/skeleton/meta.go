package skeleton

const (
	TracerNone = iota
	TracerZipkin
	TracerJaeger
)

type metaInfo struct {
	repoName    string
	baseDir     string
	projectName string

	tracer    int
	withMongo bool
}
