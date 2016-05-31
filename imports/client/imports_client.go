package client

import (
	// load the config
	_ "github.com/emccode/libstorage/imports/config"

	// load the libStorage storage driver
	_ "github.com/emccode/libstorage/drivers/storage/libstorage"

	// load the integration drivers
	_ "github.com/emccode/libstorage/drivers/integration/docker"
)
