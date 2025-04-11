package version

var (
	version   = "Unknow"
	gitCommit = "Unknow"
)

type Info struct {
	Version   string
	GitCommit string
}

func Get() Info {
	return Info{
		Version:   version,
		GitCommit: gitCommit,
	}
}
