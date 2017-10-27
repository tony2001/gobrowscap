package gobrowscap

import (
	"runtime"
	"sort"
	"sync"
)

func mergeProperties(browser *Browser, section *IniSection) *Browser {
	if browser.Parent == "" {
		browser.Parent = section.parentName
	}

	if browser.Comment == "" {
		browser.Comment = section.comment
	}

	if browser.Browser == "" {
		browser.Browser = section.browser
	}

	if browser.BrowserMaker == "" {
		browser.BrowserMaker = section.browserMaker
	}

	if browser.Version == "" {
		browser.Version = section.version
	}

	if browser.MajorVersion == "" {
		browser.MajorVersion = section.majorVersion
	}

	if browser.MinorVersion == "" {
		browser.MinorVersion = section.minorVersion
	}

	if browser.Platform == "" {
		browser.Platform = section.platform
	}

	if section.hasIsMobileDevice && !browser.HasIsMobileDevice {
		browser.IsMobileDevice = section.isMobileDevice
		browser.HasIsMobileDevice = true
	}

	if section.hasIsTablet && !browser.HasIsTablet {
		browser.IsTablet = section.isTablet
		browser.HasIsTablet = true
	}

	if section.hasCrawler && browser.HasIsCrawler {
		browser.IsCrawler = section.crawler
		browser.HasIsCrawler = true
	}

	if browser.DeviceType == "" {
		browser.DeviceType = section.deviceType
	}

	if browser.DevicePointingMethod == "" {
		browser.DevicePointingMethod = section.devicePointingMethod
	}

	if browser.PlatformVersion == "" {
		browser.PlatformVersion = section.platformVersion
	}

	if browser.BrowserType == "" {
		browser.BrowserType = section.browserType
	}

	if browser.DeviceName == "" {
		browser.DeviceName = section.deviceName
	}

	if browser.DeviceCodeName == "" {
		browser.DeviceCodeName = section.deviceCodeName
	}

	if browser.DeviceBrandName == "" {
		browser.DeviceBrandName = section.deviceBrandName
	}
	return browser
}

func SearchBrowser(iniFile *IniFile, userAgent string) (*Browser, error) {

	/* run search on all cores at once */
	goroutineBatchesNum := len(iniFile.batches)/runtime.NumCPU() + 1

	resultChan := make(chan int)
	waitFor := runtime.NumCPU()
	currentBatchIndex := 0
	for i := 0; i < goroutineBatchesNum; i++ {
		var wg sync.WaitGroup

		foundBatchIndexes := make([]int, 0)

		if i == goroutineBatchesNum-1 {
			waitFor = len(iniFile.batches) - runtime.NumCPU()*i
		}

		wg.Add(waitFor + 1)

		for j := 0; j < waitFor; j++ {
			go func(currentBatchIndex int, userAgent string) {
				defer wg.Done()
				batchMatches := iniFile.batches[currentBatchIndex].FindAllString(userAgent, -1)
				if batchMatches != nil {
					resultChan <- currentBatchIndex
				} else {
					resultChan <- -1
				}
			}(currentBatchIndex, userAgent)
			currentBatchIndex++
		}

		go func() {
			defer wg.Done()
			for k := 0; k < waitFor; k++ {
				result := <-resultChan
				if result != -1 {
					foundBatchIndexes = append(foundBatchIndexes, result)
				}
			}
		}()

		wg.Wait()

		if len(foundBatchIndexes) > 0 {
			sort.Ints(foundBatchIndexes)

			for _, batchIndex := range foundBatchIndexes {
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
					browser.Pattern = pattern.patternStr
					browser = mergeProperties(browser, section)
					for section.parentName != "" {
						section = iniFile.sections[section.parent]
						browser = mergeProperties(browser, section)
					}

					return browser, nil
				}
			}
		}
	}

	return nil, nil
}
