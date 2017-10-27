package gobrowscap

type Browser struct {
	Pattern              string
	Parent               string
	Comment              string
	Browser              string
	BrowserType          string
	BrowserMaker         string
	Platform             string
	PlatformVersion      string
	IsMobileDevice       bool
	HasIsMobileDevice    bool
	IsTablet             bool
	HasIsTablet          bool
	IsCrawler            bool
	HasIsCrawler         bool
	Version              string
	MajorVersion         string
	MinorVersion         string
	DeviceType           string
	DevicePointingMethod string
	DeviceName           string
	DeviceCodeName       string
	DeviceBrandName      string
}
