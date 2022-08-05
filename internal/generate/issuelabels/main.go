//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"
)

const (
	filename = `../../../.github/labeler-issue-labels.yml`
)

type Label struct {
	ResourceName string
	RegExp       string
}

type templateData struct {
	Labels []Label
}

func main() {
	fmt.Printf("Generating %s\n", strings.TrimPrefix(filename, "../../../"))

	productUrls := []string{
		"https://developer.atlassian.com/cloud/jira/platform/swagger-v3.v3.json",
		"https://developer.atlassian.com/cloud/confluence/swagger.v3.json",
	}

	lbs := []Label{}

	for _, v := range productUrls {
		lbs = append(lbs, getLabels(v)...)
	}

	// Add custom labels
	lbs = append(lbs, getJiraCustomLabels()...)

	td := templateData{}
	td.Labels = append(td.Labels, lbs...)

	sort.SliceStable(td.Labels, func(i, j int) bool {
		return td.Labels[i].ResourceName < td.Labels[j].ResourceName
	})

	writeTemplate(tmpl, "issuelabeler", td)
}

func getLabels(url string) []Label {
	c := http.Client{Timeout: time.Duration(2) * time.Second}
	resp, getErr := c.Get(url)
	if getErr != nil {
		log.Fatalf("error calling url (%s): %s", url, getErr)
		return nil
	}

	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatalf("error reading response body %s", readErr)
	}

	var result map[string]interface{}
	parseErr := json.Unmarshal(body, &result)
	if parseErr != nil {
		log.Fatalf("error unmarshalling JSON response %s", parseErr)
		return nil
	}
	resp.Body.Close()

	r := regexp.MustCompile(`\(apps\)`)
	r2 := regexp.MustCompile(`jira|confluence`)
	productName := r2.FindString(url)
	var tags = result["tags"].([]interface{})
	var labels []Label
	for _, value := range tags {
		// Each value is an interface{} type, that is type asserted as map[string]interface{}
		// due to nested objects in the original JSON response
		m := value.(map[string]interface{})
		rawLabel := m["name"].(string)
		ok := r.MatchString(rawLabel)
		if ok {
			continue
		}
		l := singularizeLabelSuffix(productName, rawLabel)
		labels = append(labels, l)
	}

	return labels
}

func getJiraCustomLabels() []Label {
	var customLabels []Label
	product := "jira"
	names := []string{
		"issue-field-configuration-items",
		"issue-field-configuration-schemes",
	}

	for _, n := range names {
		l := singularizeLabelSuffix(product, n)
		customLabels = append(customLabels, l)
	}

	return customLabels
}

var (
	sr  = strings.NewReplacer(" ", "", "-", "")
	sr2 = strings.NewReplacer(" - ", " ", "-", " ")
	sr3 = strings.NewReplacer(" ", "_")
	sr4 = strings.NewReplacer("__", "_")

	ies = regexp.MustCompile(`.*ies$`)                                        // match: propert[ies]
	s   = regexp.MustCompile(`.*[^aeiou]s$|.*[aeiouy][^s]es$|.*[aeiou]{2}s$`) // match: workflow[s] or module[s] or  issue[s]
	ses = regexp.MustCompile(`.*ses$`)                                        // match: statu[ses]
	es  = regexp.MustCompile(`.*[^aeiou]{2}es`)                               // match: watch[es], bush[es]
)

func singularizeLabelSuffix(product, input string) Label {
	l := Label{}

	l.ResourceName = product + "/" + strings.ToLower(sr.Replace(input))
	var str string
	if ies.MatchString(input) {
		str = strings.TrimSuffix(strings.ToLower(sr4.Replace(sr3.Replace(sr2.Replace(input)))), "ies") + "y"
	} else if s.MatchString(input) {
		str = strings.TrimSuffix(strings.ToLower(sr4.Replace(sr3.Replace(sr2.Replace(input)))), "s")
	} else if ses.MatchString(input) {
		str = strings.TrimSuffix(strings.ToLower(sr4.Replace(sr3.Replace(sr2.Replace(input)))), "es")
	} else if es.MatchString(input) {
		str = strings.TrimSuffix(strings.ToLower(sr4.Replace(sr3.Replace(sr2.Replace(input)))), "es")
	} else {
		str = strings.ToLower(sr4.Replace(sr3.Replace(sr2.Replace(input))))
	}
	l.RegExp = `((\*|-)\s*` + "`" + `?|(data|resource)\s+"?)atlassian_` + product + "_" + str + `\b`

	return l
}

func writeTemplate(body string, templateName string, td templateData) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file (%s): %s", filename, err)
	}

	tp, err := template.New(templateName).Parse(body)
	if err != nil {
		log.Fatalf("error parsing template: %s", err)
	}

	var buffer bytes.Buffer
	err = tp.Execute(&buffer, td)
	if err != nil {
		log.Fatalf("error executing template: %s", err)
	}

	if _, err := f.Write(buffer.Bytes()); err != nil {
		f.Close()
		log.Fatalf("error writing to file (%s): %s", filename, err)
	}

	if err := f.Close(); err != nil {
		log.Fatalf("error closing file (%s): %s", filename, err)
	}
}

var tmpl = `# Generated by internal/generate/issuelabels/main.go; DO NOT EDIT.
#
# ATLASSIAN Per-Resource Labeling
#
# Catch the following:
# 1. List items (* or -) with atlassian_XXX resource prefix (with or without backticks)
# 2. "data atlassian_XXX" or "resource atlassian_XXX"

{{- range .Labels }}
{{ .ResourceName }}:
  - '{{ .RegExp }}'
{{- end }}
`
