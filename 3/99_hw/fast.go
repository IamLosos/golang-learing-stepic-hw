package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hw3/models"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	// return

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	decoder := json.NewDecoder(file)

	if err != nil {
		panic(err)
	}

	seenBrowsers := []string{}
	uniqueBrowsers := 0
	var sb strings.Builder

	i := -1
	u := &models.User{}
	for decoder.More() {
		if err := decoder.Decode(u); err != nil {
			panic(err)
		}

		i++

		isAndroid := false
		isMSIE := false

		for _, browser := range u.Browsers {

			if strings.Contains(browser, "Android") {
				isAndroid = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		for _, browser := range u.Browsers {
			if strings.Contains(browser, "MSIE") {
				isMSIE = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		sb.WriteString("[")
		sb.WriteString(strconv.FormatInt(int64(i), 10))
		sb.WriteString("] ")
		sb.WriteString(u.Name)
		sb.WriteString(" <")
		sb.WriteString(strings.Replace(u.Email, "@", " [at] ", 1))
		sb.WriteString(">\n")
	}

	fmt.Fprintln(out, "found users:\n"+sb.String())
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}

// SlowSearch with problem marks
func FastSearchAsSlowSearch(out io.Writer) {
	// return

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	fileContents, err := ioutil.ReadAll(file) //TODO:OPTIMIZ
	if err != nil {
		panic(err)
	}

	r := regexp.MustCompile("@")
	seenBrowsers := []string{}
	uniqueBrowsers := 0
	foundUsers := ""

	lines := strings.Split(string(fileContents), "\n") //TODO:OPTIMIZ

	users := make([]map[string]interface{}, 0) //TODO:OPTIMIZ2
	for _, line := range lines {
		user := make(map[string]interface{}) //TODO:OPTIMIZ2
		// fmt.Printf("%v %v\n", err, line)
		err := json.Unmarshal([]byte(line), &user) //TODO:OPTIMIZ
		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	for i, user := range users {
		fmt.Println(i)

		isAndroid := false
		isMSIE := false

		browsers, ok := user["browsers"].([]interface{})
		if !ok {
			// log.Println("cant cast browsers")
			continue
		}

		for _, browserRaw := range browsers {
			browser, ok := browserRaw.(string)
			if !ok {
				// log.Println("cant cast browser to string")
				continue
			}
			if ok, err := regexp.MatchString("Android", browser); ok && err == nil { //TODO:OPTIMIZ
				isAndroid = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		for _, browserRaw := range browsers {
			browser, ok := browserRaw.(string)
			if !ok {
				// log.Println("cant cast browser to string")
				continue
			}
			if ok, err := regexp.MatchString("MSIE", browser); ok && err == nil { //TODO:OPTIMIZ
				isMSIE = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		email := r.ReplaceAllString(user["email"].(string), " [at] ")       //TODO:OPTIMIZ2
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email) //TODO:OPTIMIZ2
	}

	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}

func main() {
	fastOut := new(bytes.Buffer)
	FastSearch(fastOut)
	fastResult := fastOut.String()
	fmt.Println(fastResult)
}
