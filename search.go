package gobrowscap

import (
	//	"fmt"

	"regexp"
	"runtime"
	"sort"
	"strings"
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

var filterRegex = regexp.MustCompile(`^([A-Za-z]+)[^A-Za-z]+([A-Za-z]+)[^A-Za-z]+([A-Za-z]+).*`)

const filterSize = 3

func filterCreate(userAgent string) []string {

	matches := filterRegex.FindStringSubmatch(strings.ToLower(userAgent))
	if len(matches) == filterSize+1 /* leftmost part + 3 words */ {
		words := make([]string, filterSize)
		for i := 0; i < filterSize; i++ {
			words[i] = matches[i+1]
		}
		return words
	}
	return nil
}

func filterBatches(iniFile *IniFile, userAgent string) []int {

	filter := filterCreate(userAgent)
	if filter != nil {
		batchesToSearch := make([]int, 0)
		for i := 0; i < len(iniFile.batches); i++ {
			wordsFound := 0
			for j := 0; j < filterSize; j++ {
				if strings.Contains(iniFile.batches[i].patternStr, filter[j]) {
					wordsFound++
				} else {
					break
				}
			}
			if wordsFound == filterSize {
				batchesToSearch = append(batchesToSearch, i)
			}
		}
		return batchesToSearch
	}
	return nil
}

func searchInBatches(iniFile *IniFile, batches []*Batch, userAgent string) (*Browser, error) {
	/* run search on all cores at once */
	goroutineBatchesNum := len(batches)/runtime.NumCPU() + 1

	resultChan := make(chan int)
	waitFor := runtime.NumCPU()
	arrIndex := 0
	for i := 0; i < goroutineBatchesNum; i++ {
		var wg sync.WaitGroup

		foundBatchIndexes := make([]int, 0)

		if i == goroutineBatchesNum-1 {
			waitFor = len(batches) - runtime.NumCPU()*i
		}

		wg.Add(waitFor + 1)

		for j := 0; j < waitFor; j++ {
			go func(arrIndex int, userAgent string) {
				defer wg.Done()
				batchMatches := batches[arrIndex].regex.MatcherString(userAgent, 0).Matches()
				if batchMatches {
					resultChan <- batches[arrIndex].index
				} else {
					resultChan <- -1
				}
			}(arrIndex, userAgent)
			arrIndex++
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
				for i := batchIndex * iniFile.batchSize; i < (batchIndex+1)*iniFile.batchSize && i < len(iniFile.patterns); i++ {
					pattern := iniFile.patterns[i]
					matcher := pattern.regex.MatcherString(userAgent, 0)
					hasMatches := matcher.Matches()
					if !hasMatches {
						continue
					}

					var key int
					if matcher.Groups() == 0 {
						key = pattern.intval
					} else {
						matchString := "@"

						for m := 1; m < matcher.Groups()+1; m++ {
							if m == matcher.Groups() {
								matchString = matchString + matcher.GroupString(m)
							} else {
								matchString = matchString + matcher.GroupString(m) + "|"
							}
						}

						//						fmt.Println(pattern.patternStr)
						//						fmt.Println(matchString)

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

func SearchBrowser(iniFile *IniFile, userAgent string) (*Browser, error) {

	var filteredBatches []*Batch
	filteredBatchesIndexes := filterBatches(iniFile, userAgent)
	if filteredBatchesIndexes == nil || len(filteredBatchesIndexes) == 0 {
		return searchInBatches(iniFile, iniFile.batches, userAgent)
	} else {
		filteredBatches = make([]*Batch, len(filteredBatchesIndexes))
		for i := 0; i < len(filteredBatchesIndexes); i++ {
			filteredBatches[i] = iniFile.batches[filteredBatchesIndexes[i]]
		}

		browser, err := searchInBatches(iniFile, filteredBatches, userAgent)
		if err != nil {
			return nil, err
		}

		if browser != nil {
			return browser, nil
		}

		/* repeat with the full list */
		return searchInBatches(iniFile, iniFile.batches, userAgent)
	}

	return nil, nil
}
