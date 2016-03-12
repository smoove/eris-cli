package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func CheckoutBranch(branch string) {
	checkoutArgs := []string{"checkout", branch}

	stdOut, err := exec.Command("git", checkoutArgs...).CombinedOutput()
	if err != nil {
		log.WithField("branch", branch).Fatalf("Error checking out branch: %v", string(stdOut))
	}

	log.WithField("branch", branch).Debug("Branch checked-out")
}

func PullBranch(branch string) {
	pullArgs := []string{"pull", "origin", branch}

	stdOut, err := exec.Command("git", pullArgs...).CombinedOutput()
	if err != nil {
		log.Fatalf("Error pulling from GitHub: %v", string(stdOut))
	}

	log.WithField("branch", branch).Debug("Branch pulled successfully")
}

func InstallErisGo() {
	goArgs := []string{"install", "./cmd/eris"}

	stdOut, err := exec.Command("go", goArgs...).CombinedOutput()
	if err != nil {
		log.Fatalf("Error with go install ./cmd/eris: %v", string(stdOut))
	}

	log.Debug("Go install worked correctly")
}

func version() string {
	verArgs := []string{"version"}

	stdOut, err := exec.Command("eris", verArgs...).CombinedOutput()
	if err != nil {
		log.Fatalf("error getting version:\n%s\n", string(stdOut))
	}
	return string(stdOut)

}

// TODO break this up!!!
func DownloadLatestBinaryRelease() (string, error) {
	latestURL := "https://github.com/eris-ltd/eris-cli/releases/latest"
	resp, err := http.Get(latestURL)
	if err != nil {
		return "", fmt.Errorf("could not retrieve latest eris release at %s\nerror: %v\n", latestURL, err)
	}

	latestURL = resp.Request.URL.String()
	lastPos := strings.LastIndex(latestURL, "/")
	version := latestURL[lastPos+1:]
	platform := runtime.GOOS
	arch := runtime.GOARCH
	hostURL := "https://github.com/eris-ltd/eris-cli/releases/download/" + version + "/"
	filename := "eris_" + version[1:] + "_" + platform + "_" + arch
	fileURL := hostURL + filename

	switch platform {
	case "linux":
		filename += ".tar.gz"
	default:
		filename += ".zip"
	}

	var erisBin string
	output, err := os.Create(filename)
	// if we dont have permissions to create a file where eris cli exists, attempt to create file within HOME folder
	if err != nil {
		erisBin := filepath.Join(common.ScratchPath, "bin")
		if _, err = os.Stat(erisBin); os.IsNotExist(err) {
			err = os.MkdirAll(erisBin, 0755)
			if err != nil {
				log.Println("Error creating directory", erisBin, "Did not download binary. Exiting...")
				return "", err
			}
		}
		err = os.Chdir(erisBin)
		if err != nil {
			log.Println("Error changing directory to", erisBin, "Did not download binary. Exiting...")
			return "", err
		}
		output, err = os.Create(filename)
		if err != nil {
			log.Println("Error creating file", erisBin, "Exiting...")
			return "", err
		}
	}
	defer output.Close()

	fileResponse, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("error getting file: %v\n", err)
	}
	defer fileResponse.Body.Close()

	_, err = io.Copy(output, fileResponse.Body)
	if err != nil {
		return "", fmt.Errorf("error saving file: %v\n", err)
	}
	erisLoc, _ := exec.LookPath("eris")

	// this is hacky !!!
	if erisBin != "" {
		log.Println("downloaded eris binary", version, "for", platform, "to", erisBin, "\n Please manually move to", erisLoc)
	} else {
		log.Println("downloaded eris binary", version, "for", platform, "to", erisLoc)
	}

	var unzip string = "tar -xvf"
	if platform != "linux" {
		unzip = "unzip"
	}
	cmd := exec.Command("bin/sh", "-c", unzip, filename)
	if err := cmd.Run(); err != nil {
		return filename, fmt.Errorf("unzipping failed: %v\n", err)
	}
	return filename, nil
}
