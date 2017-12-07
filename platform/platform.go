package platform

type Platform interface {
}

type platform struct {
}

func NewPlatform() Platform {
	p := platform{}
	return &p
}
