package config

const defaultVersion = "N/A"

var (
	BuildVersion = defaultVersion
	BuildCommit  = defaultVersion
	BuildDate    = defaultVersion
)

type Version struct {
	BuildVersion string
	BuildCommit  string
	BuildDate    string
}

func GetVersion() Version {
	return Version{
		BuildVersion: BuildVersion,
		BuildCommit:  BuildCommit,
		BuildDate:    BuildDate,
	}
}
