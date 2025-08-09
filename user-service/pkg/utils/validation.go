package utils

import "regexp"

var phoneRe = regexp.MustCompile(`^[0-9+]{7,15}$`)

func ValidPhone(p string) bool { return phoneRe.MatchString(p) }
