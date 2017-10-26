## Example:
```go
package main

import "fmt"
import "gobrowscap"

func main() {

    iniFile, err := gobrowscap.LoadIniFile("/tmp/full_php_browscap.ini", 10)
    if err != nil { 
        return
    }   
    
    fmt.Println(gobrowscap.GetFileVersion(iniFile))
    
    browser, err := gobrowscap.SearchBrowser(iniFile, "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")
    fmt.Println(browser)
} 
```
