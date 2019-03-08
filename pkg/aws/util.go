package aws

import (
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
)

// GetLocalRegion gets the region ID from the instance metadata.
func GetLocalRegion() string {
	resp, err := http.Get("http://169.254.169.254/latest/meta-data/placement/availability-zone/")
	if err != nil {
		glog.Errorf("unable to get current region information, %v", err)
		return ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("cannot read response from instance metadata, %v", err)
	}

	// strip the last character from AZ to get region ID
	return string(body[0 : len(body)-1])
}
