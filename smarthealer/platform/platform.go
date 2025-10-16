package platform

type Platform int

const (
	AndroidPlatform Platform = iota
	IosPlatform
	WebPlatform
)

const (
	andStr     = "Android"
	iosStr     = "Ios"
	webStr     = "Web"
	invalidStr = "invalid platform"
)

func (p Platform) String() string {
	switch p {
	case AndroidPlatform:
		return andStr
	case IosPlatform:
		return iosStr
	case WebPlatform:
		return webStr
	default:
		return invalidStr
	}
}

func NewPlatformFromString(s string) Platform {
	switch s {
	case andStr:
		return AndroidPlatform
	case iosStr:
		return IosPlatform
	case webStr:
		return WebPlatform
	default:
		return Platform(-1)
	}
}

func (p Platform) MarshalJSON() ([]byte, error) {
	return []byte(`"` + p.String() + `"`), nil
}

func (p *Platform) UnmarshalJSON(b []byte) error {
	// Remove quotes if present
	s := string(b)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	*p = NewPlatformFromString(s)
	return nil
}
