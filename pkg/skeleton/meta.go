package skeleton

const (
	TracerZipkin = iota
	TracerJaeger
)

type metaInfo struct {
	repoName    string
	baseDir     string
	projectName string

	tracer    int
	withMongo bool
}
