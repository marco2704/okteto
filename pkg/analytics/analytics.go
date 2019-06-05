package analytics

import (
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/dukex/mixpanel"
	"github.com/okteto/app/cli/pkg/config"
	"github.com/okteto/app/cli/pkg/log"
	"github.com/okteto/app/cli/pkg/okteto"
)

const (
	mixpanelToken = "92fe782cdffa212d8f03861fbf1ea301"

	createDatabaseEvent = "Create Database"
	upEvent             = "Up"
	downEvent           = "Down"
	loginEvent          = "Login"
	createEvent         = "Create Manifest"
	execEvent           = "Exec"
)

var (
	mixpanelClient mixpanel.Mixpanel
	machineID      string
)

func init() {
	c := &http.Client{
		Timeout: time.Second * 5,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}

	mixpanelClient = mixpanel.NewFromClient(c, mixpanelToken, "")

	var err error
	machineID, err = machineid.ProtectedID("okteto")
	if err != nil {
		log.Debugf("failed to generate a machine id")
		machineID = "na"
	}
}

// TrackCreate sends a tracking event to mixpanel when the user creates a manifest
func TrackCreate(language, image, version string) {
	track(createEvent, version, image)
}

// TrackUp sends a tracking event to mixpanel when the user activates a development environment
func TrackUp(image, version string) {
	track(upEvent, version, image)
}

// TrackExec sends a tracking event to mixpanel when the user runs the exec command
func TrackExec(image, version string) {
	track(execEvent, version, image)
}

// TrackDown sends a tracking event to mixpanel when the user deactivates a development environment
func TrackDown(image, version string) {
	track(downEvent, version, image)
}

// TrackLogin sends a tracking event to mixpanel when the user logs in
func TrackLogin(version string) {
	track(loginEvent, version, "")
}

func track(event, version, image string) {
	if isEnabled() {
		e := &mixpanel.Event{
			Properties: map[string]interface{}{
				"os":                runtime.GOOS,
				"version":           version,
				"$referring_domain": okteto.GetURL(),
				"image":             image,
				"machine_id":        machineID,
			},
		}

		trackID := okteto.GetUserID()
		if len(trackID) == 0 {
			trackID = machineID
		}

		if err := mixpanelClient.Track(machineID, event, e); err != nil {
			log.Infof("Failed to send analytics: %s", err)
		}
	} else {
		log.Debugf("not sending event for %s", event)
	}
}

func getFlagPath() string {
	return filepath.Join(config.GetHome(), ".noanalytics")
}

// Disable disables analytics
func Disable() error {
	var _, err = os.Stat(getFlagPath())
	if os.IsNotExist(err) {
		var file, err = os.Create(getFlagPath())
		if err != nil {
			return err
		}

		defer file.Close()
	}

	return nil
}

// Enable enables analytics
func Enable() error {
	var _, err = os.Stat(getFlagPath())
	if os.IsNotExist(err) {
		return nil
	}

	return os.Remove(getFlagPath())
}

func isEnabled() bool {
	if _, err := os.Stat(getFlagPath()); !os.IsNotExist(err) {
		return false
	}

	return true
}
