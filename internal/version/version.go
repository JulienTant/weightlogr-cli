package version

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

type BuildInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

func Info() BuildInfo {
	return BuildInfo{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
	}
}
