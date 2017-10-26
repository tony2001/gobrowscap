package gobrowscap

type Browser struct {
	pattern              string
	parent               string
	comment              string
	browser              string
	browserType          string
	browserMaker         string
	platform             string
	platformVersion      string
	isMobileDevice       bool
	hasIsMobileDevice    bool
	isTablet             bool
	hasIsTablet          bool
	isCrawler            bool
	hasIsCrawler         bool
	version              string
	majorVersion         string
	minorVersion         string
	deviceType           string
	devicePointingMethod string
	deviceName           string
	deviceCodeName       string
	deviceBrandName      string
}
