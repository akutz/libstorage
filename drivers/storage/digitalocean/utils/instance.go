// +build !libstorage_storage_driver libstorage_storage_driver_digitalocean

package utils

import (
	"io/ioutil"
	"net/http"

	"github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/libstorage/drivers/storage/digitalocean"
)

const (
	metadataBase   = "169.254.169.254"
	metadataURL    = "http://" + metadataBase + "/metadata/v1"
	metadataID     = metadataURL + "/id"
	metadataRegion = metadataURL + "/region"
	metadataName   = metadataURL + "/hostname"
)

// InstanceID gets the instance information from the droplet
func InstanceID(ctx types.Context) (*types.InstanceID, error) {

	id, err := getURL(ctx, metadataID)
	if err != nil {
		return nil, err
	}

	region, err := getURL(ctx, metadataRegion)
	if err != nil {
		return nil, err
	}

	name, err := getURL(ctx, metadataName)
	if err != nil {
		return nil, err
	}

	return &types.InstanceID{
		ID:     id,
		Driver: digitalocean.Name,
		Fields: map[string]string{
			digitalocean.InstanceIDFieldRegion: region,
			digitalocean.InstanceIDFieldName:   name,
		},
	}, nil
}

func getURL(ctx types.Context, url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := doRequest(ctx, req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	id, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(id), nil
}
