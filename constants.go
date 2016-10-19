package bond

type MongoDBEdition string

const (
	Enterprise        MongoDBEdition = "enterprise"
	CommunityTargeted                = "targeted"
	Base                             = "base"
)

type MongoDBArch string

const (
	ZSeries MongoDBArch = "s390x"
	POWER               = "ppc64le"
	AMD64               = "x86_64"
	X86                 = "i686"
)
