package gobrowscap

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
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
	if browser, ok := SearchBrowser(FILE, TEST_USER_AGENT); ok != nil {
		t.Error("Browser not found")
	} else if browser.Browser != "Chrome" {
		t.Errorf("Expected Chrome but got %q", browser.Browser)
	} else if browser.Platform != "MacOSX" {
		t.Errorf("Expected MacOSX but got %q", browser.Platform)
	} else if browser.Version != "37.0" {
		t.Errorf("Expected 37.0 but got %q", browser.Version)
	} else if browser.IsCrawler != false {
		t.Errorf("Expected false but got %q", browser.IsCrawler)
	}
}

func TestGetBrowserIPhone(t *testing.T) {
	if browser, ok := SearchBrowser(FILE, TEST_IPHONE_AGENT); ok != nil {
		t.Error("Browser not found")
	} else if browser.DeviceName != "iPhone" {
		t.Errorf("Expected iPhone but got %q", browser.DeviceName)
	} else if browser.Platform != "iOS" {
		t.Errorf("Expected iOS but got %q", browser.Platform)
	} else if browser.IsMobileDevice != true {
		t.Errorf("Expected true but got %t", browser.IsMobileDevice)
	}
}

func TestGetBrowserYandex(t *testing.T) {
	if browser, ok := SearchBrowser(FILE, TEST_YANDEX_AGENT); ok != nil {
		t.Error("Browser not found")
	} else if browser.Browser != "Yandex Browser" {
		t.Errorf("Expected Yandex Browser but got %q", browser.Browser)
	} else if browser.IsCrawler != false {
		t.Errorf("Expected false but got %t", browser.IsCrawler)
	}
}

func TestGetBrowserAndroid(t *testing.T) {
	if browser, ok := SearchBrowser(FILE, TEST_ANDROID_AGENT); ok != nil {
		t.Error("Browser not found")
	} else if browser.DeviceName != "Galaxy Note" {
		t.Errorf("Expected Galaxy Note but got %q", browser.DeviceName)
	} else if browser.Platform != "Android" {
		t.Errorf("Expected Android but got %q", browser.Platform)
	} else if browser.Browser != "Chrome" {
		t.Errorf("Expected Chrome but got %q", browser.Browser)
	} else if browser.IsCrawler != false {
		t.Errorf("Expected false but got %t", browser.IsCrawler)
	}
}

func TestGetBrowserMobileFirefox(t *testing.T) {
	if browser, ok := SearchBrowser(FILE, TEST_MOBILE_FIREFOX); ok != nil {
		t.Error("Browser not found")
	} else if browser.DeviceName != "general Mobile Phone" {
		t.Errorf("Expected general Mobile Phone but got %q", browser.DeviceName)
	} else if browser.Platform != "Android" {
		t.Errorf("Expected Android but got %q", browser.Platform)
	} else if browser.Browser != "Firefox" {
		t.Errorf("Expected Firefox but got %q", browser.Browser)
	} else if browser.IsCrawler != false {
		t.Errorf("Expected false but got %t", browser.IsCrawler)
	}
}

func TestGetBrowserIssues(t *testing.T) {
	// https://github.com/digitalcrab/browscap_go/issues/4
	ua := "Mozilla/5.0 (iPad; CPU OS 5_0_1 like Mac OS X) AppleWebKit/534.46 (KHTML, like Gecko) Version/5.1 Mobile/9A405 Safari/7534.48.3"
	if browser, ok := SearchBrowser(FILE, ua); ok != nil {
		t.Error("Browser not found")
	} else if browser.DeviceType != "Tablet" {
		t.Errorf("Expected tablet %q", browser.DeviceType)
	}
}
func TestLastVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	version := GetFileVersion(FILE)
	if version == "" {
		t.Fatalf("Version not found")
	}
	//t.Logf("Last version is %q, current version: %q", version, InitializedVersion())
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

		browser, ok := SearchBrowser(FILE, uas[idx])
		if ok != nil {
			b.Errorf("User agent not recognized: %s", uas[idx])
		}
		if browser == nil {
			b.Errorf("User agent not recognized: %s", uas[idx])
		}
	}
}
