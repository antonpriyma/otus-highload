package utils

import "net/http"

func GetCookieByName(cookies []*http.Cookie, name string) string {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.Value
		}
	}

	return ""
}
