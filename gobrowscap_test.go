package gobrowscap

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TEST_INI_FILE       = "./test-data/full_php_browscap.ini"
	TEST_USER_AGENT     = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/37.0.1062.110 Safari/537.36"
	TEST_IPHONE_AGENT   = "Mozilla/5.0 (iPhone; U; CPU iPhone OS 4_3_2 like Mac OS X; en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8H7 Safari/6533.18.5"
	TEST_YANDEX_AGENT   = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/30.0.1599.12785 YaBrowser/13.12.1599.12785 Safari/537.36"
	TEST_ANDROID_AGENT  = "Mozilla/5.0 (Linux; Android 4.0.4; GT-N7000 Build/IMM76D) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.93 Mobile Safari/537.36"
	TEST_MOBILE_FIREFOX = "Mozilla/5.0 (Android 4.1.1; Mobile; rv:55.0) Gecko/55.0 Firefox/55.0"
)

var FILE *IniFile

func TestMain(m *testing.M) {
	var err error
	if FILE, err = LoadIniFile(TEST_INI_FILE, 10); err != nil {
		os.Exit(-1)
	}
	m.Run()
}

/*
func TestLoadIniFile(t *testing.T) {
	var err error
	if FILE, err = LoadIniFile(TEST_INI_FILE, 0); err != nil {
		t.Fatalf("%v", err)
	}
}
*/
func TestSearchBrowser(t *testing.T) {
	browser, err := SearchBrowser(FILE, TEST_USER_AGENT)
	require.NoError(t, err)

	assert.Equal(t, "Chrome", browser.Browser)
	assert.Equal(t, "MacOSX", browser.Platform)
	assert.Equal(t, "37.0", browser.Version)
	assert.False(t, browser.IsCrawler)
}

func TestGetBrowserIPhone(t *testing.T) {
	browser, err := SearchBrowser(FILE, TEST_IPHONE_AGENT)
	require.NoError(t, err)

	assert.Equal(t, "Safari", browser.Browser)
	assert.Equal(t, "iOS", browser.Platform)
	assert.True(t, browser.IsMobileDevice)
}

func TestGetBrowserYandex(t *testing.T) {
	browser, err := SearchBrowser(FILE, TEST_YANDEX_AGENT)
	require.NoError(t, err)

	assert.Equal(t, "Yandex Browser", browser.Browser)
	assert.False(t, browser.IsCrawler)
}

func TestGetBrowserAndroid(t *testing.T) {
	browser, err := SearchBrowser(FILE, TEST_ANDROID_AGENT)
	require.NoError(t, err)

	assert.Equal(t, "Galaxy Note", browser.DeviceName)
	assert.Equal(t, "Android", browser.Platform)
	assert.Equal(t, "Chrome", browser.Browser)
	assert.False(t, browser.IsCrawler)
}

func TestGetBrowserMobileFirefox(t *testing.T) {
	browser, err := SearchBrowser(FILE, TEST_MOBILE_FIREFOX)
	require.NoError(t, err)

	assert.Equal(t, "general Mobile Phone", browser.DeviceName)
	assert.Equal(t, "Android", browser.Platform)
	assert.Equal(t, "Firefox", browser.Browser)
	assert.False(t, browser.IsCrawler)
}

func TestGetBrowserIssues(t *testing.T) {
	// https://github.com/digitalcrab/browscap_go/issues/4
	ua := "Mozilla/5.0 (iPad; CPU OS 5_0_1 like Mac OS X) AppleWebKit/534.46 (KHTML, like Gecko) Version/5.1 Mobile/9A405 Safari/7534.48.3"
	browser, err := SearchBrowser(FILE, ua)
	require.NoError(t, err)

	assert.Equal(t, "Tablet", browser.DeviceType)
}
func TestLastVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	version := GetFileVersion(FILE)
	assert.NotEmpty(t, version)
}

func BenchmarkInit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LoadIniFile(TEST_INI_FILE, 100)
	}
}

func BenchmarkSearchBrowser(b *testing.B) {
	data, err := ioutil.ReadFile("test-data/user_agents_sample.txt")
	if err != nil {
		b.Error(err)
	}

	uas := strings.Split(strings.TrimSpace(string(data)), "\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(uas)

		browser, err := SearchBrowser(FILE, uas[idx])
		require.NoError(b, err)
		assert.NotNil(b, browser)
	}
}
