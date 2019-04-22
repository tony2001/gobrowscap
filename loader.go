/*

This code is based on two other projects:
1. https://github.com/digitalcrab/browscap_go
2. https://github.com/GaretJax/phpbrowscap/

*/

package gobrowscap

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

type TmpPattern struct {
	intval   int
	first    string
	matches  map[int][]string
	parent   int
	position int
}

type DeduplicatedPattern struct {
	intval   int
	matches  map[string]int
	parent   int
	position int
}

type Pattern struct {
	priority    int
	position    int
	length      int
	shortLength int
	patternStr  string
	regex       *pcre.Regexp
	intval      int
	matches     map[string]int
}

type Batch struct {
	regex      *pcre.Regexp
	patternStr string
	index      int
}

type IniFile struct {
	patterns  []*Pattern
	sections  map[int]*IniSection
	batches   []*Batch
	batchSize int
	version   string
}

var (
	// Ini
	sEmpty   = []byte{}     // empty signal
	nComment = []byte{'#'}  // number signal
	sComment = []byte{';'}  // semicolon signal
	sStart   = []byte{'['}  // section start signal
	sEnd     = []byte{']'}  // section end signal
	sEqual   = []byte{'='}  // equal signal
	sQuote1  = []byte{'"'}  // quote " signal
	sQuote2  = []byte{'\''} // quote ' signal

	versionSection = "GJK_Browscap_Version"
	versionKey     = "Version"
)

type IniSection struct {
	parent               int
	parentName           string
	comment              string
	browser              string
	browserMaker         string
	version              string
	majorVersion         string
	minorVersion         string
	platform             string
	platformVersion      string
	isMobileDevice       bool
	hasIsMobileDevice    bool
	isTablet             bool
	hasIsTablet          bool
	crawler              bool
	hasCrawler           bool
	deviceType           string
	devicePointingMethod string
	browserType          string
	deviceName           string
	deviceCodeName       string
	deviceBrandName      string
}

func regexUnquote(quotedRegex string, matches []string) string { /* {{{ */
	replaceMap := make(map[string]string)
	replaceMap[`\.`] = `\?`
	replaceMap[`\\`] = `\`
	replaceMap[`\+`] = `+`
	replaceMap[`\[`] = `[`
	replaceMap[`\^`] = `^`
	replaceMap[`\]`] = `]`
	replaceMap[`\$`] = `$`
	replaceMap[`\(`] = `(`
	replaceMap[`\)`] = `)`
	replaceMap[`\{`] = `{`
	replaceMap[`\{`] = `}`
	replaceMap[`\=`] = `=`
	replaceMap[`\!`] = `!`
	replaceMap[`\<`] = `<`
	replaceMap[`\>`] = `>`
	replaceMap[`\|`] = `|`
	replaceMap[`\:`] = `:`
	replaceMap[`\-`] = `-`
	replaceMap[`.*`] = `*`
	replaceMap[`\?`] = `?`
	replaceMap[`\.`] = `.`

	for old, new := range replaceMap {
		quotedRegex = strings.Replace(quotedRegex, old, new, -1)
	}

	var resultStr string
	if len(quotedRegex) > 4 {
		// this affects sorting of the patterns
		// but I have no idea why is this needed at all
		resultStr = quotedRegex[2 : len(quotedRegex)-2]
	} else {
		return ""
	}

	for i := 0; i < len(matches); i++ {
		resultStr = strings.Replace(resultStr, `(\d)`, matches[i], 1)
	}
	return resultStr
}

/* }}} */

func parseBoolValue(fieldName string, value string, lineNum int) (bool, error) { /* {{{ */
	if value == "true" {
		return true, nil
	} else if value == "false" {
		return false, nil
	}
	return false, fmt.Errorf("invalid value for %s: expected true/false, got '%s' on line %d", fieldName, value, lineNum)
}

/* }}} */

func parseSectionValues(section *IniSection, key string, value string, lineNum int) (*IniSection, error) { /* {{{ */
	switch key {
	case "Parent":
		section.parentName = value
	case "Comment":
		section.comment = value
	case "Browser":
		section.browser = value
	case "Browser_Maker":
		section.browserMaker = value
	case "Version":
		section.version = value /* can contain non-numeric symbols*/
	case "MajorVer":
		section.majorVersion = value
	case "MinorVer":
		section.minorVersion = value
	case "Platform":
		section.platform = value
	case "Platform_Version":
		section.platformVersion = value
	case "isMobileDevice":
		isMobileDevice, err := parseBoolValue(key, value, lineNum)
		if err != nil {
			return nil, err
		}
		section.isMobileDevice = isMobileDevice
		section.hasIsMobileDevice = true
	case "isTablet":
		isTablet, err := parseBoolValue(key, value, lineNum)
		if err != nil {
			return nil, err
		}
		section.isTablet = isTablet
		section.hasIsTablet = true
	case "Crawler":
		crawler, err := parseBoolValue(key, value, lineNum)
		if err != nil {
			return nil, err
		}
		section.crawler = crawler
		section.hasCrawler = true
	case "Device_Type":
		section.deviceType = value
	case "Device_Pointing_Method":
		section.devicePointingMethod = value
	case "Browser_Type":
		section.browserType = value
	case "Device_Name":
		section.deviceName = value
	case "Device_Code_Name":
		section.deviceCodeName = value
	case "Device_Brand_Name":
		section.deviceBrandName = value
	default:
		/* ignore the others */
	}
	return section, nil
}

/* }}} */

func parseIniFile(path string) (string, map[int]string, map[int]*IniSection, error) { /* {{{ */
	file, err := os.Open(path)
	if err != nil {
		return "", nil, nil, err
	}
	defer file.Close()

	buf := bufio.NewReader(file)

	sectionName := ""
	sectionNum := 0
	version := ""

	sectionMap := make(map[string]int)
	sections := make(map[int]*IniSection)

	lineNum := 0
	isVersionSection := false
	/* parse INI and create maps of sections and properties */
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return "", nil, nil, err
			}
		}
		lineNum++

		// Empty line
		if bytes.Equal(sEmpty, line) {
			continue
		}

		// Trim
		line = bytes.TrimSpace(line)

		// Empty line
		if bytes.Equal(sEmpty, line) {
			continue
		}

		// Comment line
		if bytes.HasPrefix(line, nComment) || bytes.HasPrefix(line, sComment) {
			continue
		}

		// Section line
		if bytes.HasPrefix(line, sStart) && bytes.HasSuffix(line, sEnd) {
			sectionName = string(line[1 : len(line)-1])
			if sectionName == versionSection {
				isVersionSection = true
			} else {
				isVersionSection = false
				sectionMap[sectionName] = sectionNum
				sections[sectionNum] = new(IniSection)
				sectionNum++
			}
			continue
		}

		// Key => Value
		kv := bytes.SplitN(line, sEqual, 2)

		// Parse Key
		keyb := bytes.TrimSpace(kv[0])

		// Parse Value
		valb := bytes.TrimSpace(kv[1])
		if bytes.HasPrefix(valb, sQuote1) {
			valb = bytes.Trim(valb, `"`)
		}
		if bytes.HasPrefix(valb, sQuote2) {
			valb = bytes.Trim(valb, `'`)
		}

		key := string(keyb)
		val := string(valb)

		if isVersionSection {
			if key == versionKey {
				version = val
			}
			continue
		}

		section, err := parseSectionValues(sections[sectionNum-1], key, val, lineNum)
		if err != nil {
			return "", nil, nil, err
		}
		sections[sectionNum-1] = section
	}

	if len(sectionMap) != len(sections) {
		return "", nil, nil, fmt.Errorf("parse failure: number of len(sectionMap) != len(sectionProperties). duplicate section name?")
	}

	var sectionMapInverted = make(map[int]string)
	for section, index := range sectionMap {
		sectionMapInverted[index] = section

		parentName := sections[index].parentName
		if parentName != "" {
			parentIndex, ok := sectionMap[parentName]
			if ok == true {
				sections[index].parent = parentIndex
			} else {
				return "", nil, nil, fmt.Errorf("unknown Parent value specified (not present in the section names): '%s'", parentName)
			}
		}
	}
	return version, sectionMapInverted, sections, nil
}

/* }}} */

func mergeMap(a map[int]string, b map[int]string) map[int]string { /* {{{ */
	for k, v := range b {
		_, ok := a[k]
		if ok {
			continue
		}
		a[k] = v
	}
	return a
}

/* }}} */

func diffAssocArrMapToArr(a []string, b map[int]string) []string { /* {{{ */
	resultArr := make([]string, 0)

	for i := 0; i < len(a); i++ {
		vb, ok := b[i]
		if ok {
			if vb != a[i] {
				resultArr = append(resultArr, a[i])
			}
		} else {
			resultArr = append(resultArr, a[i])
		}
	}

	for kb, vb := range b {
		if kb >= len(a) {
			resultArr = append(resultArr, vb)
		}
	}

	return resultArr
}

/* }}} */

func diffAssocArr(a []string, b []string) map[int]string { /* {{{ */
	resultMap := make(map[int]string)
	for kb, vb := range b {
		if kb < len(a) {
			va := a[kb]
			if va == vb {
				continue
			}
			resultMap[kb] = va
		} else {
			resultMap[kb] = vb
		}
	}

	for ka, va := range a {
		if ka > len(b) {
			resultMap[ka] = va
		}
	}
	return resultMap
}

/* }}} */

func deduplicateCompressionPattern(matches map[int][]string, pattern string) (map[string]int, string) { /* {{{ */
	matchesCopy := make(map[int][]string, len(matches))

	minIndex := 0
	for index, value := range matches {
		if minIndex == 0 || minIndex > index {
			minIndex = index
		}
		matchesCopy[index] = value
	}

	firstMatch := matchesCopy[minIndex]
	delete(matchesCopy, minIndex)

	differences := make(map[int]string)
	for _, value := range matchesCopy {
		tmp := diffAssocArr(firstMatch, value)
		differences = mergeMap(differences, tmp)
	}

	identical := make(map[int]string)
	for i := 0; i < len(firstMatch); i++ {
		_, ok := differences[i]
		if !ok {
			identical[i] = firstMatch[i]
		}
	}

	prepared_matches := make(map[string]int)
	for index, match := range matches {
		tmp := diffAssocArrMapToArr(match, identical)

		key := "@" + strings.Join(tmp, "|")
		prepared_matches[key] = index
	}

	pattern_parts := strings.Split(pattern, `(\d)`)

	for i := 0; i < len(pattern_parts); i++ {
		val, ok := identical[i]
		if ok {
			pattern_parts[i+1] = pattern_parts[i] + val + pattern_parts[i+1]
			/* since there's no way to delete it without recreating the array and
			empty string is a valid value here, let's just use a special marker */
			pattern_parts[i] = "###"
		}
	}

	pattern_parts_clean := make([]string, 0)
	for i := 0; i < len(pattern_parts); i++ {
		if pattern_parts[i] != "###" {
			pattern_parts_clean = append(pattern_parts_clean, pattern_parts[i])
		}
	}

	resultPattern := strings.Join(pattern_parts_clean, `(\d)`)
	return prepared_matches, resultPattern
}

/* }}} */

func deduplicatePatterns(patterns map[string]*TmpPattern) map[string]*DeduplicatedPattern { /* {{{ */
	resultMap := make(map[string]*DeduplicatedPattern)
	for key, value := range patterns {
		result := new(DeduplicatedPattern)
		if value.intval == 0 {
			if len(value.matches) == 1 && value.first != "" {
				key = value.first
				for match_key, _ := range value.matches {
					result.intval = match_key
				}
			} else {
				result.matches, key = deduplicateCompressionPattern(value.matches, key)
			}
		} else {
			result.intval = value.intval
		}
		result.position = value.position
		result.parent = value.parent
		resultMap[key] = result
	}
	return resultMap
}

/* }}} */

func compileAndAddBatchRegex(batchesArr []*Batch, patternStr string, batchIndex int) ([]*Batch, error) { /* {{{ */
	regex, err := pcre.Compile(patternStr, pcre.CASELESS)
	if err != nil {
		return nil, fmt.Errorf("%s", err.String())
	}
	batch := new(Batch)
	batch.regex = &regex
	batch.patternStr = patternStr
	batch.index = batchIndex
	batchesArr[batchIndex] = batch
	return batchesArr, nil
}

/* }}} */

func createRegexpBatches(patterns []*Pattern, batchSize int) ([]*Batch, error) { /* {{{ */
	var err error
	batchIndex := 0
	numInBatch := 1
	batches := make([]*Batch, len(patterns)/batchSize+1)
	batchStr := "^"
	for i := 0; i < len(patterns); i++ {

		batchStr = batchStr + "(?:" + strings.ToLower(patterns[i].patternStr) + ")"

		if numInBatch < batchSize {
			batchStr = batchStr + "|"
			numInBatch++
			continue
		}

		batchStr = batchStr + "$"

		batches, err = compileAndAddBatchRegex(batches, batchStr, batchIndex)
		if err != nil {
			return nil, err
		}
		batchIndex++

		numInBatch = 1
		batchStr = "(?i)^"
	}

	if numInBatch < batchSize {
		batchStr = batchStr + "$"

		batches, err = compileAndAddBatchRegex(batches, batchStr, batchIndex)
		if err != nil {
			return nil, err
		}
	}

	return batches, nil
}

/* }}} */

func processIniSections(sectionMap map[int]string, sections map[int]*IniSection) map[string]*TmpPattern { /* {{{ */
	tmpPatterns := make(map[string]*TmpPattern)

	for i := 0; i < len(sectionMap); i++ {
		userAgent := sectionMap[i]
		section := sections[i]
		/* looks like Comment is only present for very high-level sections, which are used as Parent's for others */
		if section.comment == "" || strings.Contains(userAgent, "*") || strings.Contains(userAgent, "?") {
			pattern := regexp.QuoteMeta(strings.ToLower(userAgent))
			pattern = strings.Replace(pattern, `:`, `\:`, -1)   //just to make sure the results are
			pattern = strings.Replace(pattern, `-`, `\-`, -1)   //ordered the same way as in PHP
			pattern = strings.Replace(pattern, `\*`, `.*`, -1)  //
			pattern = strings.Replace(pattern, `\?`, `.`, -1)   //
			pattern = strings.Replace(pattern, `\x`, `\\x`, -1) //the \\x replacement is a fix for "Der gro\xdfe BilderSauger 2.00u" user agent match" (c)

			regex, _ := regexp.Compile("\\d")
			matches := regex.FindAllString(pattern, -1)

			if len(matches) == 0 {
				p := new(TmpPattern)
				p.intval = i
				p.position = i
				tmpPatterns[pattern] = p
			} else {
				compressedPattern := regex.ReplaceAllString(pattern, `(\d)`)

				_, ok := tmpPatterns[compressedPattern]
				if ok == false {
					p := new(TmpPattern)
					p.first = pattern
					p.position = i
					p.matches = make(map[int][]string)
					p.parent = 0

					if section.parent > 0 {
						p.parent = section.parent
					}
					tmpPatterns[compressedPattern] = p
				}

				tmpPatterns[compressedPattern].matches[i] = matches
			}
		}
	}
	return tmpPatterns
}

/* }}} */

func LoadIniFile(path string, batchSize int) (*IniFile, error) { /* {{{ */
	version, sectionMap, sections, err := parseIniFile(path)
	if err != nil {
		return nil, err
	}

	tmpPatterns := processIniSections(sectionMap, sections)

	patterns := deduplicatePatterns(tmpPatterns)

	i := 0
	readyPatterns := make([]*Pattern, len(patterns))
	for patternString, patternObj := range patterns {
		regex, err := pcre.Compile("^"+patternString+"$", pcre.CASELESS)
		if err != nil {
			return nil, fmt.Errorf("Failed to compile regexp: %s, err: %s", patternString, err)
		}

		decodedPattern := regexUnquote(patternString, nil)
		decodedPattern = strings.Replace(decodedPattern, `(\d)`, "0", -1)

		ready := new(Pattern)
		ready.priority = 1
		if decodedPattern == "*" {
			/* "*" has to be the last one */
			ready.priority = 2
		}

		ready.regex = &regex
		ready.intval = patternObj.intval
		ready.position = patternObj.position
		ready.matches = patternObj.matches
		ready.length = len(decodedPattern)

		/* this also affects resulting sort order */
		shortPattern := strings.Replace(decodedPattern, "*", "", -1)
		shortPattern = strings.Replace(shortPattern, "?", "", -1)
		ready.shortLength = len(shortPattern)
		ready.patternStr = patternString

		readyPatterns[i] = ready
		i++
	}

	sort.Slice(readyPatterns, func(i, j int) bool {
		a := readyPatterns[i]
		b := readyPatterns[j]
		if a.priority != b.priority {
			return a.priority < b.priority
		}
		if a.length != b.length {
			return a.length > b.length
		}
		if a.shortLength != b.shortLength {
			return a.shortLength > b.shortLength
		}
		if a.position != b.position {
			return a.position < b.position
		}
		return true
	})

	batches, err := createRegexpBatches(readyPatterns, batchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to compile batch regex: %s", err)
	}

	iniFile := new(IniFile)
	iniFile.patterns = readyPatterns
	iniFile.sections = sections
	iniFile.batches = batches
	iniFile.batchSize = batchSize
	iniFile.version = version

	return iniFile, nil
}

/* }}} */
