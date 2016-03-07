package files

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/ipfs"
)

func GetFiles(do *definitions.Do) error {
	ensureRunning()

	dirBool := checkPath(do.Path)

	if dirBool {
		if err := importDirectory(do); err != nil {
			return err
		}
		//get like you put dir
	} else {
		if err := importFile(do.Name, do.Path); err != nil {
			return err
		}
	}
	do.Result = "success"
	return nil
}

func PutFiles(do *definitions.Do) error {
	ensureRunning()

	if err := checkGatewayFlag(do); err != nil {
		return err
	}

	//check if do.Name is dir or file ...
	f, err := os.Stat(do.Name)
	if err != nil {
		return err
	}

	if f.IsDir() {
		//can't use gateway - check & throw err
		if err := exportDirectory(do); err != nil {
			return err
		}

	} else {
		hash, err := exportFile(do.Name, do.Gateway)
		if err != nil {
			return err
		}
		do.Result = hash
	}
	return nil
}

func exportDirectory(do *definitions.Do) error {
	log.WithField("dir", do.Name).Warn("Adding contents of a directory")

	// path to dir on host
	do.Source = do.Name
	// path to dest in cont (doesn't exist, need to make it)
	// will be removed later
	do.Destination = filepath.Join(ErisContainerRoot, "scratch", "data", do.Source)
	do.Name = "ipfs"

	do.Operations.Interactive = false
	do.Operations.PublishAllPorts = true
	do.Operations.Args = []string{"mkdir", "-p", do.Destination}

	if err := services.ExecService(do); err != nil {
		return err
	}

	do.Operations.Args = nil
	do.Operations.PublishAllPorts = false
	if err := data.ImportData(do); err != nil {
		return err
	}

	ip := new(bytes.Buffer)
	config.GlobalConfig.Writer = ip

	do.Operations.Interactive = false
	do.Operations.PublishAllPorts = true
	do.Operations.Args = []string{"NetworkSettings.IPAddress"}

	if err := services.InspectService(do); err != nil {
		return err
	}
	api := fmt.Sprintf("/ip4/%s/tcp/5001", util.TrimString(ip.String()))

	do.Operations.Interactive = false
	do.Operations.PublishAllPorts = true
	do.Operations.Args = []string{"ipfs", "add", "-r", do.Destination, "--api", api}

	out := new(bytes.Buffer)
	config.GlobalConfig.Writer = out

	if err := services.ExecService(do); err != nil {
		return err
	}
	log.Warn("Directory object added succesfully")
	log.Warn(out.String())

	return nil
}

func importDirectory(do *definitions.Do) error {
	log.WithFields(log.Fields{
		"hash": do.Name,
		"path": do.Path,
	}).Warn("Getting a directory")
	hash := do.Name

	// path to dir on host
	do.Destination = do.Path
	// path to source in cont (doesn't exist, need to make it)
	// exec'ing ipfs get will save files in Root hash named dir
	do.Source = filepath.Join(ErisContainerRoot, "scratch", "data", do.Path)
	do.Name = "ipfs"

	do.Operations.Interactive = false
	do.Operations.PublishAllPorts = true
	do.Operations.Args = []string{"mkdir", "-p", do.Source}

	if err := services.ExecService(do); err != nil {
		return err
	}

	ip := new(bytes.Buffer)
	config.GlobalConfig.Writer = ip

	do.Operations.Interactive = false
	do.Operations.PublishAllPorts = true
	do.Operations.Args = []string{"NetworkSettings.IPAddress"}

	if err := services.InspectService(do); err != nil {
		return err
	}
	api := fmt.Sprintf("/ip4/%s/tcp/5001", util.TrimString(ip.String()))

	do.Operations.Interactive = false
	do.Operations.PublishAllPorts = true
	//will save to /home/eris/.eris in hash dir.
	do.Operations.Args = []string{"ipfs", "get", hash, "--api", api}

	out := new(bytes.Buffer)
	config.GlobalConfig.Writer = out

	if err := services.ExecService(do); err != nil {
		return err
	}
	log.Warn(out.String())

	do.Operations.Args = []string{"mv", filepath.Join(ErisContainerRoot, hash), do.Source}

	if err := services.ExecService(do); err != nil {
		return err
	}

	//get src/dest right
	do.Operations.Args = nil
	do.Operations.PublishAllPorts = false
	if err := data.ExportData(do); err != nil {
		return err
	}

	log.Warn("Directory object getted succesfully.")
	//log.Warn(out.String())
	return nil

}
func PinFiles(do *definitions.Do) error {
	ensureRunning()
	log.WithFields(log.Fields{
		"file": do.Name,
		"path": do.Path,
	}).Debug("Pinning a file")
	hash, err := pinFile(do.Name)
	if err != nil {
		return err
	}
	do.Result = hash
	return nil
}

func CatFiles(do *definitions.Do) error {
	ensureRunning()
	log.WithFields(log.Fields{
		"file": do.Name,
		"path": do.Path,
	}).Debug("Dumping the contents of a file")
	hash, err := catFile(do.Name)
	if err != nil {
		return err
	}
	do.Result = hash
	return nil
}

func ListFiles(do *definitions.Do) error {
	ensureRunning()
	log.WithFields(log.Fields{
		"file": do.Name,
		"path": do.Path,
	}).Debug("Listing an object")
	hash, err := listFile(do.Name)
	if err != nil {
		return err
	}
	do.Result = hash
	return nil
}

func ManagePinned(do *definitions.Do) error {
	ensureRunning()
	if do.Rm && do.Hash != "" {
		return fmt.Errorf("Either remove a file by hash or all of them\n")
	}

	if do.Rm {
		log.Info("Removing all cached files")
		hashes, err := rmAllPinned()
		if err != nil {
			return err
		}
		do.Result = hashes
	} else if do.Hash != "" {
		log.WithField("hash", do.Hash).Info("Removing from cache")
		hashes, err := rmPinnedByHash(do.Hash)
		if err != nil {
			return err
		}
		do.Result = hashes
	} else {
		log.Debug("Listing files pinned locally")
		hash, err := listPinned()
		if err != nil {
			return err
		}
		do.Result = hash
	}
	return nil
}

func importFile(hash, fileName string) error {
	var err error

	log.WithFields(log.Fields{
		"from hash": hash,
		"to path":   fileName,
	}).Debug("Importing a file")

	if log.GetLevel() > 0 {
		err = ipfs.GetFromIPFS(hash, fileName, "", os.Stdout)
	} else {
		err = ipfs.GetFromIPFS(hash, fileName, "", bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return err
	}
	return nil
}

func exportFile(fileName, gateway string) (string, error) {
	var hash string
	var err error

	log.WithFields(log.Fields{
		"file":    fileName,
		"gateway": gateway,
	}).Debug("Adding a file")

	if log.GetLevel() > 0 {
		hash, err = ipfs.SendToIPFS(fileName, gateway, os.Stdout)
	} else {
		hash, err = ipfs.SendToIPFS(fileName, gateway, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}

	return hash, nil
}

func pinFile(fileHash string) (string, error) {
	var hash string
	var err error

	if log.GetLevel() > 0 {
		hash, err = ipfs.PinToIPFS(fileHash, os.Stdout)
	} else {
		hash, err = ipfs.PinToIPFS(fileHash, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}
	return hash, nil
}

func catFile(fileHash string) (string, error) {
	var hash string
	var err error
	if log.GetLevel() > 0 {
		hash, err = ipfs.CatFromIPFS(fileHash, os.Stdout)
	} else {
		hash, err = ipfs.CatFromIPFS(fileHash, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}
	return hash, nil
}

func listFile(objectHash string) (string, error) {
	var hash string
	var err error
	if log.GetLevel() > 0 {
		hash, err = ipfs.ListFromIPFS(objectHash, os.Stdout)
	} else {
		hash, err = ipfs.ListFromIPFS(objectHash, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}
	return hash, nil
}

func listPinned() (string, error) {
	var hash string
	var err error
	if log.GetLevel() > 0 {
		hash, err = ipfs.ListPinnedFromIPFS(os.Stdout)
	} else {
		hash, err = ipfs.ListPinnedFromIPFS(bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}
	return hash, nil
}

func rmAllPinned() (string, error) {
	hashList, err := listPinned()
	if err != nil {
		return "", err
	}

	hashArray := strings.Split(hashList, "\n")
	result := make([]string, len(hashArray))
	for i, hash := range hashArray {
		result[i], err = rmPinnedByHash(hash)
		if err != nil {
			return "", err
		}
	}
	hashes := strings.Join(result, "\n")
	return hashes, nil
}

func rmPinnedByHash(hash string) (string, error) {
	var err error
	if log.GetLevel() > 0 {
		hash, err = ipfs.RemovePinnedFromIPFS(hash, os.Stdout)
	} else {
		hash, err = ipfs.RemovePinnedFromIPFS(hash, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}
	return hash, nil
}

//---------------------------------------------------------
// helpers

func ensureRunning() {
	doNow := definitions.NowDo()
	doNow.Name = "ipfs"
	err := services.EnsureRunning(doNow)
	if err != nil {
		fmt.Printf("Failed to ensure IPFS is running: %v", err)
		return
	}
	log.Info("IPFS is running")
}

func checkGatewayFlag(do *definitions.Do) error {
	if do.Gateway != "" {
		_, err := url.Parse(do.Gateway)
		if err != nil {
			return fmt.Errorf("Invalid gateway URL provided %v\n", err)
		}
		log.WithField("gateway", do.Gateway).Debug("Posting to")
	} else {
		log.Debug("Posting to gateway.ipfs.io")
	}
	return nil
}

func checkPath(path string) bool {
	dirBool := false
	thing := strings.Split(path, ".")
	if len(thing) == 1 {
		log.Warn("No file extension detected, assuming directory.")
		return true
	} else {
		log.Warn("File extension detected, assuming file.")
		return false
	}
	return dirBool
}
