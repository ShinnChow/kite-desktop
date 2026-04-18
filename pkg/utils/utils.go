package utils

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	kiteBasePlaceholder = "__KITE_BASE__"
	umamiScriptSrc      = "https://umami.eryajf.net/script.js"
	umamiWebsiteID      = "8317012e-c8ab-4b59-bc86-2e708ceac202"
)

func InjectAnalytics(htmlContent string) string {
	analyticsScript := fmt.Sprintf(
		`<script defer src="%s" data-website-id="%s" data-auto-track="false" data-tag="desktop" data-domains="127.0.0.1,localhost" data-exclude-search="true" data-exclude-hash="true" data-do-not-track="true"></script>`,
		umamiScriptSrc,
		umamiWebsiteID,
	)

	re := regexp.MustCompile(`<head>`)
	return re.ReplaceAllString(htmlContent, "<head>\n    "+analyticsScript)
}

func InjectKiteBase(htmlContent string, base string, analyticsEnabled bool) string {
	assetBase := base
	if assetBase == "/" {
		assetBase = ""
	}

	htmlContent = strings.ReplaceAll(htmlContent, kiteBasePlaceholder, html.EscapeString(assetBase))

	baseScript := fmt.Sprintf(
		`<script>window.__dynamic_base__=%q;window.__kite_analytics_enabled__=%t;</script>`,
		assetBase,
		analyticsEnabled,
	)
	re := regexp.MustCompile(`<head>`)
	return re.ReplaceAllString(htmlContent, "<head>\n    "+baseScript)
}

func RandomString(length int) string {
	return rand.String(length)
}

func ToEnvName(input string) string {
	s := input
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ToUpper(s)
	return s
}

func GetImageRegistryAndRepo(image string) (string, string) {
	image = strings.SplitN(image, ":", 2)[0]
	parts := strings.Split(image, "/")
	if len(parts) == 1 {
		return "", "library/" + parts[0]
	}
	if len(parts) > 1 {
		if strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":") {
			return parts[0], strings.Join(parts[1:], "/")
		}
		return "", strings.Join(parts, "/")
	}
	return "", image
}
