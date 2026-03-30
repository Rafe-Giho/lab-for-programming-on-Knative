package runtimecatalog

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Runtime struct {
	Language string `json:"language"`
	Version  string `json:"version"`
	Image    string `json:"image"`
}

type LanguageSummary struct {
	Key      string   `json:"key"`
	Label    string   `json:"label"`
	Versions []string `json:"versions"`
}

type Catalog struct {
	runtimes map[string]Runtime
}

func New(images map[string]string) Catalog {
	runtimes := make(map[string]Runtime, len(images))
	for runtimeKey, image := range images {
		language, version := parseRuntimeKey(runtimeKey)
		runtimes[key(language, version)] = Runtime{
			Language: language,
			Version:  version,
			Image:    image,
		}
	}
	return Catalog{runtimes: runtimes}
}

func (c Catalog) Find(language, version string) (Runtime, error) {
	runtime, ok := c.runtimes[key(language, version)]
	if !ok {
		return Runtime{}, fmt.Errorf("unsupported runtime: %s %s", language, version)
	}
	return runtime, nil
}

func (c Catalog) PythonVersions() []string {
	return c.VersionsFor("python")
}

func (c Catalog) VersionsFor(language string) []string {
	versions := make([]string, 0, len(c.runtimes))
	for _, runtime := range c.runtimes {
		if runtime.Language == language {
			versions = append(versions, runtime.Version)
		}
	}
	sort.Slice(versions, func(i, j int) bool {
		return versionLess(versions[i], versions[j])
	})
	return versions
}

func (c Catalog) Languages() []string {
	seen := make(map[string]struct{}, len(c.runtimes))
	languages := make([]string, 0, len(c.runtimes))
	for _, runtime := range c.runtimes {
		if _, ok := seen[runtime.Language]; ok {
			continue
		}
		seen[runtime.Language] = struct{}{}
		languages = append(languages, runtime.Language)
	}
	sort.Slice(languages, func(i, j int) bool {
		return languageRank(languages[i]) < languageRank(languages[j])
	})
	return languages
}

func (c Catalog) Supported() map[string][]string {
	supported := make(map[string][]string)
	for _, language := range c.Languages() {
		supported[language] = c.VersionsFor(language)
	}
	return supported
}

func (c Catalog) Summaries() []LanguageSummary {
	languages := c.Languages()
	summaries := make([]LanguageSummary, 0, len(languages))
	for _, language := range languages {
		summaries = append(summaries, LanguageSummary{
			Key:      language,
			Label:    displayLanguage(language),
			Versions: c.VersionsFor(language),
		})
	}
	return summaries
}

func key(language, version string) string {
	return language + ":" + version
}

func parseRuntimeKey(runtimeKey string) (language, version string) {
	parts := strings.SplitN(runtimeKey, ":", 2)
	if len(parts) == 1 {
		return "python", runtimeKey
	}
	return parts[0], parts[1]
}

func displayLanguage(language string) string {
	switch language {
	case "python":
		return "Python"
	case "java":
		return "Java"
	case "c":
		return "C"
	case "cpp":
		return "C++"
	default:
		return strings.ToUpper(language)
	}
}

func languageRank(language string) int {
	switch language {
	case "python":
		return 0
	case "java":
		return 1
	case "c":
		return 2
	case "cpp":
		return 3
	default:
		return 99
	}
}

func versionLess(left, right string) bool {
	leftParts := strings.Split(left, ".")
	rightParts := strings.Split(right, ".")

	for idx := 0; idx < len(leftParts) && idx < len(rightParts); idx++ {
		leftValue, leftErr := strconv.Atoi(leftParts[idx])
		rightValue, rightErr := strconv.Atoi(rightParts[idx])

		if leftErr != nil || rightErr != nil {
			return left < right
		}
		if leftValue != rightValue {
			return leftValue < rightValue
		}
	}

	return len(leftParts) < len(rightParts)
}
