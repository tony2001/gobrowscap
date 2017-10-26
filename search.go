package gobrowscap

func mergeProperties(browser *Browser, section *IniSection) *Browser {
	if browser.parent == "" {
		browser.parent = section.parentName
	}

	if browser.comment == "" {
		browser.comment = section.comment
	}

	if browser.browser == "" {
		browser.browser = section.browser
	}

	if browser.browserMaker == "" {
		browser.browserMaker = section.browserMaker
	}

	if browser.version == "" {
		browser.version = section.version
	}

	if browser.majorVersion == "" {
		browser.majorVersion = section.majorVersion
	}

	if browser.minorVersion == "" {
		browser.minorVersion = section.minorVersion
	}

	if browser.platform == "" {
		browser.platform = section.platform
	}

	if section.hasIsMobileDevice && !browser.hasIsMobileDevice {
		browser.isMobileDevice = section.isMobileDevice
		browser.hasIsMobileDevice = true
	}

	if section.hasIsTablet && !browser.hasIsTablet {
		browser.isTablet = section.isTablet
		browser.hasIsTablet = true
	}

	if section.hasCrawler && browser.hasIsCrawler {
		browser.isCrawler = section.crawler
		browser.hasIsCrawler = true
	}

	if browser.deviceType == "" {
		browser.deviceType = section.deviceType
	}

	if browser.devicePointingMethod == "" {
		browser.devicePointingMethod = section.devicePointingMethod
	}

	if browser.platformVersion == "" {
		browser.platformVersion = section.platformVersion
	}

	if browser.browserType == "" {
		browser.browserType = section.browserType
	}

	if browser.deviceName == "" {
		browser.deviceName = section.deviceName
	}

	if browser.deviceCodeName == "" {
		browser.deviceCodeName = section.deviceCodeName
	}

	if browser.deviceBrandName == "" {
		browser.deviceBrandName = section.deviceBrandName
	}
	return browser
}

func SearchBrowser(iniFile *IniFile, userAgent string) (*Browser, error) {

	for batchIndex := 0; batchIndex < len(iniFile.batches); batchIndex++ {
		batchMatches := iniFile.batches[batchIndex].FindAllString(userAgent, -1)
		if batchMatches == nil {
			continue
		}

		for i := batchIndex * iniFile.batchSize; i < (batchIndex+1)*iniFile.batchSize; i++ {
			pattern := iniFile.patterns[i]
			matches := pattern.regex.FindStringSubmatch(userAgent)
			if matches == nil {
				continue
			}

			var key int
			if len(matches) == 1 {
				key = pattern.intval
			} else {
				matchString := "@"
				for i := 1; i < len(matches); i++ {
					if i == len(matches)-1 {
						matchString = matchString + matches[i]
					} else {
						matchString = matchString + matches[i] + "|"
					}
				}

				var ok bool
				key, ok = pattern.matches[matchString]
				if !ok {
					/* partial match, continue search */
					continue
				}
			}

			section := iniFile.sections[key]

			browser := new(Browser)
			browser.pattern = pattern.patternStr
			browser = mergeProperties(browser, section)
			for section.parentName != "" {
				section = iniFile.sections[section.parent]
				browser = mergeProperties(browser, section)
			}

			return browser, nil
		}
	}

	return nil, nil
}
