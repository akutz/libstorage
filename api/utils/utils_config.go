package utils

import (
	"fmt"
	"strings"

	"github.com/akutz/gofig"
	"github.com/emccode/libstorage/api/types"
)

func isSet(
	config gofig.Config,
	key types.ConfigKey,
	roots ...string) bool {

	keySz := key.String()

	for _, r := range roots {
		rk := strings.Replace(keySz, "libstorage.", fmt.Sprintf("%s.", r), 1)
		if config.IsSet(rk) {
			return true
		}
	}

	if config.IsSet(keySz) {
		return true
	}

	return false
}

func getString(
	config gofig.Config,
	key types.ConfigKey,
	roots ...string) string {

	var (
		val   string
		keySz = key.String()
	)

	for _, r := range roots {
		rk := strings.Replace(keySz, "libstorage.", fmt.Sprintf("%s.", r), 1)
		if val = config.GetString(rk); val != "" {
			return val
		}
	}

	val = config.GetString(key)
	if val != "" {
		return val
	}

	return ""
}

func getBool(
	config gofig.Config,
	key types.ConfigKey,
	roots ...string) bool {

	keySz := key.String()

	for _, r := range roots {
		rk := strings.Replace(keySz, "libstorage.", fmt.Sprintf("%s.", r), 1)
		if config.IsSet(rk) {
			return config.GetBool(rk)
		}
	}

	if config.IsSet(keySz) {
		return config.GetBool(keySz)
	}

	return false
}
