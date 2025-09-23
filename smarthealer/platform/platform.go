package platform

type Platform int

const (
	AndroidPlatform Platform = iota
	IosPlatform
	WebPlatform
)

func (p Platform) String() string {
	switch p {
	case AndroidPlatform:
		return "Android"
	case IosPlatform:
		return "Ios"
	case WebPlatform:
		return "Web"
	default:
		return "invalid platform"
	}
}
