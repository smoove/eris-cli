package chains

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	ini "github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/list"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/tests"
	"github.com/eris-ltd/eris-cli/util"
	ver "github.com/eris-ltd/eris-cli/version"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	logger "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

var (
	erisDir   = filepath.Join(os.TempDir(), "eris")
	chainName = "test-chain"
)

func TestMain(m *testing.M) {
	log.SetFormatter(logger.ErisFormatter{})

	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	tests.IfExit(tests.TestsInit("chain"))
	mockChainDefinitionFile(chainName)

	exitCode := m.Run()

	tests.IfExit(tests.TestsTearDown())

	os.Exit(exitCode)
}

func TestListAllChainsRunning(t *testing.T) {
	if err := mockChainDefinitionFile("test-chain-list1"); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	do := def.NowDo()
	do.Known = true
	do.Existing = false
	do.Running = true
	do.Quiet = true
	if err := list.ListAll(do, "chains"); err != nil {
		t.Fatalf("expected list to succeed, got %v", err)
	}

	if strings.Contains(do.Result, "test-chain-list1") {
		t.Fatalf("expected the chain not found")
	}
}

func TestListAllChainsKnown(t *testing.T) {
	if err := mockChainDefinitionFile("test-chain-list2"); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	do := def.NowDo()
	do.Known = true
	do.Existing = false
	do.Running = false
	do.Quiet = true
	if err := list.ListAll(do, "chains"); err != nil {
		t.Fatalf("expected list to succeed, got %v", err)
	}

	if !strings.Contains(do.Result, "test-chain-list2") {
		t.Fatalf("expected the chain to be found")
	}
}

func TestChainGraduate(t *testing.T) {
	do := def.NowDo()
	do.Name = chainName
	if err := GraduateChain(do); err != nil {
		t.Fatalf("expected chain to graduate, got %v", err)
	}

	srvDef, err := loaders.LoadServiceDefinition(chainName, false)
	if err != nil {
		t.Fatalf("expected service definition to be loaded")
	}

	if image := path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_DB); srvDef.Service.Image != image {
		t.Fatalf("bad image on graduate, expected %s, got: %s", image, srvDef.Service.Image)
	}

	if srvDef.Service.Command != loaders.ErisChainStart {
		t.Fatalf("improper service command on graduate, expected %s, got %s", loaders.ErisChainStart, srvDef.Service.Command)
	}

	if !srvDef.Service.AutoData {
		t.Fatalf("improper service autodata value on graduate, expected %t, got %t", true, srvDef.Service.AutoData)
	}

	if len(srvDef.Dependencies.Services) != 1 {
		t.Fatalf("improper service deps on graduate, expected: [%q], got %s", "keys", srvDef.Dependencies.Services)
	}
}

func TestLoadChainDefinition(t *testing.T) {
	// [pv]: this test belongs to the loaders package.
	var err error
	chain, err := loaders.LoadChainDefinition(chainName, false)
	if err != nil {
		t.Fatalf("expected chain definition to be loaded, got %v", err)
	}

	if chain.Service.Name != chainName {
		t.Fatalf("improper service name on load, expected %s, got %s", chainName, chain.Service.Name)
	}

	if !chain.Service.AutoData {
		t.Fatalf("data_container not properly read on load, expected false")
	}

	if chain.Operations.DataContainerName == "" {
		t.Fatalf("data_container_name not set")
	}
}

func TestStartChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	start(t, chainName)
	if n := util.HowManyContainersRunning(chainName, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 chain container, got %v", n)
	}
	if n := util.HowManyContainersExisting(chainName, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 data containers, got %v", n)
	}

	kill(t, chainName)
	if n := util.HowManyContainersRunning(chainName, def.TypeChain); n != 0 {
		t.Fatalf("expecting 0 chain container, got %v", n)
	}
	if n := util.HowManyContainersExisting(chainName, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data containers, got %v", n)
	}
}

func TestRestartChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	create(t, chainName)
	if n := util.HowManyContainersRunning(chainName, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 chain container, got %v", n)
	}
	if n := util.HowManyContainersExisting(chainName, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 data containers, got %v", n)
	}

	kill(t, chainName)
	if n := util.HowManyContainersRunning(chainName, def.TypeChain); n != 0 {
		t.Fatalf("expecting 0 chain container, got %v", n)
	}
	if n := util.HowManyContainersExisting(chainName, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data containers, got %v", n)
	}

	start(t, chainName)
	if n := util.HowManyContainersRunning(chainName, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 chain container, got %v", n)
	}
	if n := util.HowManyContainersExisting(chainName, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 data containers, got %v", n)
	}

	kill(t, chainName)
	if n := util.HowManyContainersRunning(chainName, def.TypeChain); n != 0 {
		t.Fatalf("expecting 0 chain container, got %v", n)
	}
	if n := util.HowManyContainersExisting(chainName, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data containers, got %v", n)
	}
}

func TestExecChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	create(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.Operations.Args = []string{"ls", common.ErisContainerRoot}
	buf, err := ExecChain(do)
	if err != nil {
		t.Fatalf("expected chain to execute, got %v", err)
	}

	if dir := "chains"; !strings.Contains(buf.String(), dir) {
		t.Fatalf("expected to find %q dir in eris root", dir)
	}
}

func TestExecChainBadCommandLine(t *testing.T) {
	defer tests.RemoveAllContainers()

	create(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.Operations.Args = strings.Fields("bad command line")
	if _, err := ExecChain(do); err == nil {
		t.Fatalf("expected chain to fail")
	}
}

func TestCatChainLocalConfig(t *testing.T) {
	buf := new(bytes.Buffer)
	config.GlobalConfig.Writer = buf

	do := def.NowDo()
	do.Name = chainName
	do.Type = "toml"
	if err := CatChain(do); err != nil {
		t.Fatalf("expected getting a local config to succeed, got %v", err)
	}

	if buf.String() != tests.FileContents(filepath.Join(erisDir, "chains", chainName+".toml")) {
		t.Fatalf("expected the local config file to match, got %v", buf.String())
	}
}

func TestCatChainContainerConfig(t *testing.T) {
	defer tests.RemoveAllContainers()

	buf := new(bytes.Buffer)
	config.GlobalConfig.Writer = buf

	do := def.NowDo()
	do.ConfigFile = filepath.Join(common.ChainsPath, "default", "config.toml")
	do.Name = chainName
	do.Operations.PublishAllPorts = true
	if err := NewChain(do); err != nil {
		t.Fatalf("expected a new chain to be created, got %v", err)
	}

	do.Type = "config"
	if err := CatChain(do); err != nil {
		t.Fatalf("expected getting a local config to succeed, got %v", err)
	}

	if !strings.Contains(buf.String(), "defaulttester.com") {
		t.Fatalf("expected the config file to contain an expected string, got %v", buf.String())
	}
}

func TestCatChainContainerGenesis(t *testing.T) {
	defer tests.RemoveAllContainers()

	buf := new(bytes.Buffer)
	config.GlobalConfig.Writer = buf

	do := def.NowDo()
	do.ConfigFile = filepath.Join(common.ChainsPath, "default", "config.toml")
	do.Name = chainName
	do.Operations.PublishAllPorts = true
	if err := NewChain(do); err != nil {
		t.Fatalf("expected a new chain to be created, got %v", err)
	}

	do.Type = "genesis"
	if err := CatChain(do); err != nil {
		t.Fatalf("expected getting a local config to succeed, got %v", err)
	}

	if !strings.Contains(buf.String(), "accounts") || !strings.Contains(buf.String(), "validators") {
		t.Fatalf("expected the genesis file to contain expected strings, got %v", buf.String())
	}
}

func TestChainsNewDirGenesis(t *testing.T) {
	defer tests.RemoveAllContainers()

	const chain = "test-dir-gen"

	do := def.NowDo()
	do.Name = chain
	do.Operations.PublishAllPorts = true
	if err := NewChain(do); err != nil {
		t.Fatalf("expected a new chain to be created, got %v", err)
	}

	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/chains/%s/genesis.json", chain)}
	if out := exec(t, chain, args); !strings.Contains(out, chain) {
		t.Fatalf("expected chain_id to be equal to chain name in genesis file, got %v", out)
	}
}

func TestChainsNewConfig(t *testing.T) {
	defer tests.RemoveAllContainers()

	const chain = "test-config-new"

	do := def.NowDo()
	do.Name = chain
	do.ConfigFile = filepath.Join(common.ChainsPath, "default", "config.toml")
	// deprecated. functionality to be removed in 0.11.4
	// do.CSV = filepath.Join(common.ChainsPath, "default", "genesis.json")
	do.Operations.PublishAllPorts = true
	if err := NewChain(do); err != nil {
		t.Fatalf("expected to create a new chain, got %v", err)
	}

	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/chains/%s/config.toml", chain)}
	if out := exec(t, chain, args); !strings.Contains(out, "defaulttester.com") {
		t.Fatalf("expected the config file to contain an expected string, got %v", out)
	}
}

// chains new should import the priv_validator.json (available in mint form)
// into eris-keys (available in eris form) so it can be used by the rest
// of the platform
func TestChainsNewKeysImported(t *testing.T) {
	defer tests.RemoveAllContainers()

	const chain = "test-config-keys"

	do := def.NowDo()
	do.Name = chain
	do.Operations.PublishAllPorts = true
	if err := NewChain(do); err != nil {
		t.Fatalf("expected a new chain to be created, got %v", err)
	}

	if n := util.HowManyContainersRunning(chain, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 chain container running, got %v", n)
	}

	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/keys/data/%s/%s", ini.DefaultAddr, ini.DefaultAddr)}
	if out := exec(t, chain, args); !strings.Contains(out, ini.DefaultAddr) {
		t.Fatalf("expected to find keys in container, got %v", out)
	}
}

func TestLogsChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	create(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.Follow = false
	do.Tail = "all"
	if err := LogsChain(do); err != nil {
		t.Fatalf("failed to fetch container logs")
	}
}

func TestUpdateChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	create(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.Pull = false
	do.Operations.PublishAllPorts = true
	if err := UpdateChain(do); err != nil {
		t.Fatalf("expected chain to update, got %v", err)
	}

	if n := util.HowManyContainersRunning(chainName, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 chain container running, got %v", n)
	}
}

func TestInspectChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	create(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.Operations.Args = []string{"name"}
	if err := InspectChain(do); err != nil {
		t.Fatalf("expected chain to be inspected, got %v", err)
	}
}

func TestRenameChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	const (
		chain   = "hichain"
		rename1 = "niahctset"
	)

	do := def.NowDo()
	do.Name = chain
	if err := NewChain(do); err != nil {
		t.Fatalf("expected a new chain to be created, got %v", err)
	}

	if n := util.HowManyContainersRunning(chain, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 chain container, got %v", n)
	}
	if n := util.HowManyContainersExisting(chain, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 data containers, got %v", n)
	}

	do = def.NowDo()
	do.Name = chain
	do.NewName = rename1
	if err := RenameChain(do); err != nil {
		t.Fatalf("expected chain to be renamed #1, got %v", err)
	}

	if n := util.HowManyContainersRunning(chain, def.TypeChain); n != 0 {
		t.Fatalf("expecting 0 chain container, got %v", n)
	}
	if n := util.HowManyContainersExisting(chain, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data containers, got %v", n)
	}

	if n := util.HowManyContainersRunning(rename1, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 chain container, got %v", n)
	}
	if n := util.HowManyContainersExisting(rename1, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 data containers, got %v", n)
	}

	do = def.NowDo()
	do.Name = rename1
	do.NewName = chainName
	if err := RenameChain(do); err != nil {
		t.Fatalf("expected chain to be renamed #2, got %v", err)
	}

	if n := util.HowManyContainersRunning(rename1, def.TypeChain); n != 0 {
		t.Fatalf("expecting 0 chain container, got %v", n)
	}
	if n := util.HowManyContainersExisting(rename1, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data containers, got %v", n)
	}

	if n := util.HowManyContainersRunning(chainName, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 chain container, got %v", n)
	}
	if n := util.HowManyContainersExisting(chainName, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 data containers, got %v", n)
	}
}

func TestRmChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	mockChainDefinitionFile(chainName)

	create(t, chainName)

	do := def.NowDo()
	do.Operations.Args, do.Rm, do.RmD = []string{"keys"}, true, true
	if err := services.KillService(do); err != nil {
		t.Fatalf("expected service to be stopped, got %v", err)
	}

	do = def.NowDo()
	do.Name, do.Rm, do.RmD = chainName, false, false
	if err := KillChain(do); err != nil {
		t.Fatalf("expected chain to be stopped, got %v", err)
	}
	if n := util.HowManyContainersExisting(chainName, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 chain containers, got %v", n)
	}

	do = def.NowDo()
	do.Name = chainName
	do.RmD = true
	if err := RmChain(do); err != nil {
		t.Fatalf("expected chain to be removed, got %v", err)
	}
	if n := util.HowManyContainersExisting(chainName, def.TypeChain); n != 0 {
		t.Fatalf("expecting 0 chain containers, got %v", n)
	}
}

func TestServiceLinkNoChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	if err := tests.FakeServiceDefinition(erisDir, "fake", `
chain = "$chain:fake"

[service]
name = "fake"
image = "`+path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_IPFS)+`"
data_container = true
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	do := def.NowDo()
	do.Operations.Args = []string{"fake"}
	if err := services.StartService(do); err == nil {
		t.Fatalf("expect start service to fail, got nil")
	}
}

func TestServiceLinkBadChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	if err := tests.FakeServiceDefinition(erisDir, "fake", `
chain = "$chain:fake"

[service]
name = "fake"
image = "`+path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_IPFS)+`"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	do := def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.ChainName = "non-existent-chain"
	if err := services.StartService(do); err == nil {
		t.Fatalf("expect start service to fail, got nil")
	}
}

func TestServiceLinkBadChainWithoutChainInDefinition(t *testing.T) {
	defer tests.RemoveAllContainers()

	create(t, chainName)

	if err := tests.FakeServiceDefinition(erisDir, "fake", `
[service]
name = "fake"
image = "`+path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_IPFS)+`"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	do := def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.ChainName = "non-existent-chain"

	// [pv]: is this a bug? the service which doesn't have a
	// "chain" in its definition file doesn't care about linking at all.
	if err := services.StartService(do); err != nil {
		t.Fatalf("expect service to start, got %v", err)
	}

	if n := util.HowManyContainersRunning("fake", def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting("fake", def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data containers, got %v", n)
	}
}

func TestServiceLink(t *testing.T) {
	defer tests.RemoveAllContainers()

	const chain = "test-chain-link"

	do := def.NowDo()
	do.Name = chain
	if err := NewChain(do); err != nil {
		t.Fatalf("could not start a new chain, got %v", err)
	}

	if err := tests.FakeServiceDefinition(erisDir, "fake", `
chain = "$chain:fake"

[service]
name = "fake"
image = "`+path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_KEYS)+`"
data_container = false
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if n := util.HowManyContainersExisting(chain, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 test chain containers, got %v", n)
	}

	if n := util.HowManyContainersExisting("fake", def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting("fake", def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data containers, got %v", n)
	}

	do = def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.ChainName = chain
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	if n := util.HowManyContainersRunning("fake", def.TypeService); n != 1 {
		t.Fatalf("expecting 1 fake service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting("fake", def.TypeData); n != 0 {
		t.Fatalf("expecting 0 fake data containers, got %v", n)
	}

	links := tests.Links("fake", def.TypeService)
	if len(links) != 1 || !strings.Contains(links[0], chain) {
		t.Fatalf("expected service be linked to a test chain, got %v", links)
	}
}

func TestServiceLinkWithDataContainer(t *testing.T) {
	defer tests.RemoveAllContainers()

	const chain = "test-chain-data-container"

	do := def.NowDo()
	do.Name = chain
	if err := NewChain(do); err != nil {
		t.Fatalf("could not start a new chain, got %v", err)
	}

	if err := tests.FakeServiceDefinition(erisDir, "fake", `
chain = "$chain:fake"

[service]
name = "fake"
image = "`+path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_IPFS)+`"
data_container = true
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if n := util.HowManyContainersExisting(chain, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 test chain containers, got %v", n)
	}

	if n := util.HowManyContainersExisting("fake", def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting("fake", def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data containers, got %v", n)
	}

	do = def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.ChainName = chain
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	if n := util.HowManyContainersRunning("fake", def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting("fake", def.TypeData); n != 1 {
		t.Fatalf("expecting 1 data containers, got %v", n)
	}

	links := tests.Links("fake", def.TypeService)
	if len(links) != 1 || !strings.Contains(links[0], chain) {
		t.Fatalf("expected service be linked to a test chain, got %v", links)
	}
}

func TestServiceLinkLiteral(t *testing.T) {
	defer tests.RemoveAllContainers()

	const chain = "test-chain-literal"

	do := def.NowDo()
	do.Name = chain
	if err := NewChain(do); err != nil {
		t.Fatalf("could not start a new chain, got %v", err)
	}

	if err := tests.FakeServiceDefinition(erisDir, "fake", `
chain = "`+chain+`:fake"

[service]
name = "fake"
image = "`+path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_KEYS)+`"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if n := util.HowManyContainersExisting(chain, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 test chain containers, got %v", n)
	}

	if n := util.HowManyContainersExisting("fake", def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting("fake", def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data containers, got %v", n)
	}

	do = def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.ChainName = chain
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	if n := util.HowManyContainersRunning("fake", def.TypeService); n != 1 {
		t.Fatalf("expecting 1 fake service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting("fake", def.TypeData); n != 0 {
		t.Fatalf("expecting 0 fake data containers, got %v", n)
	}

	links := tests.Links("fake", def.TypeService)
	if len(links) != 1 || !strings.Contains(links[0], chain) {
		t.Fatalf("expected service be linked to a test chain, got %v", links)
	}
}

func TestServiceLinkBadLiteral(t *testing.T) {
	defer tests.RemoveAllContainers()

	const chain = "test-chain-bad-literal"

	do := def.NowDo()
	do.Name = chain
	if err := NewChain(do); err != nil {
		t.Fatalf("could not start a new chain, got %v", err)
	}

	if err := tests.FakeServiceDefinition(erisDir, "fake", `
chain = "blah-blah:blah"

[service]
name = "fake"
image = "`+path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_IPFS)+`"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if n := util.HowManyContainersExisting(chain, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 test chain containers, got %v", n)
	}

	do = def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.ChainName = chain
	// [pv]: probably a bug. Bad literal chain link in a definition
	// file doesn't affect the service start. Links is not nil.
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	links := tests.Links("fake", def.TypeService)
	if len(links) != 1 || !strings.Contains(links[0], chain) {
		t.Fatalf("expected service be linked to a test chain, got %v", links)
	}
}

func TestServiceLinkChainedService(t *testing.T) {
	defer tests.RemoveAllContainers()

	const chain = "test-chained-service"

	if err := tests.FakeServiceDefinition(erisDir, "fake", `
chain = "$chain:fake"

[service]
name = "fake"
image = "`+path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_KEYS)+`"

[dependencies]
services = [ "sham" ]
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if err := tests.FakeServiceDefinition(erisDir, "sham", `
chain = "$chain:sham"

[service]
name = "sham"
image = "`+path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_KEYS)+`"
data_container = true
`); err != nil {
		t.Fatalf("can't create a sham service definition: %v", err)
	}

	if n := util.HowManyContainersExisting(chain, def.TypeChain); n != 0 {
		t.Fatalf("expecting 0 test chain containers, got %v", n)
	}

	do := def.NowDo()
	do.Name = chain
	if err := NewChain(do); err != nil {
		t.Fatalf("could not start a new chain, got %v", err)
	}

	if n := util.HowManyContainersExisting(chain, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 test chain containers, got %v", n)
	}

	do = def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.ChainName = chain
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	if n := util.HowManyContainersRunning("fake", def.TypeService); n != 1 {
		t.Fatalf("expecting 1 fake service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting("fake", def.TypeData); n != 0 {
		t.Fatalf("expecting 0 fake data containers, got %v", n)
	}
	if n := util.HowManyContainersRunning("sham", def.TypeService); n != 1 {
		t.Fatalf("expecting 1 sham service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting("sham", def.TypeData); n != 1 {
		t.Fatalf("expecting 1 sham data containers, got %v", n)
	}

	// [pv]: second service doesn't reference the chain.
	links := tests.Links("fake", def.TypeService)
	if len(links) != 2 || (!strings.Contains(links[1], chain) && !strings.Contains(links[0], chain)) {
		t.Fatalf("expected service be linked to a test chain, got %v", links)
	}
}

func TestServiceLinkKeys(t *testing.T) {
	defer tests.RemoveAllContainers()

	const chain = "chain-test-keys"

	do := def.NowDo()
	do.Name = chain
	if err := NewChain(do); err != nil {
		t.Fatalf("could not start a new chain, got %v", err)
	}

	if n := util.HowManyContainersExisting(chain, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 test chain containers, got %v", n)
	}

	do = def.NowDo()
	do.Operations.Args = []string{"keys"}
	do.ChainName = chain
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	if n := util.HowManyContainersExisting("keys", def.TypeService); n != 1 {
		t.Fatalf("expecting 1 test chain containers, got %v", n)
	}

	links := tests.Links("keys", def.TypeService)
	if len(links) != 0 {
		t.Fatalf("expected service links be empty, got %v", links)
	}
}

func create(t *testing.T, chain string) {
	do := def.NowDo()
	do.ConfigFile = filepath.Join(common.ChainsPath, "default", "config.toml")
	do.Name = chain
	do.Operations.PublishAllPorts = true
	if err := NewChain(do); err != nil {
		t.Fatalf("expected a new chain to be created, got %v", err)
	}
}

func start(t *testing.T, chain string) {
	do := def.NowDo()
	do.Name = chain
	do.Operations.PublishAllPorts = true
	if err := StartChain(do); err != nil {
		t.Fatalf("starting chain %v failed: %v", chain, err)
	}
}

func kill(t *testing.T, chain string) {
	do := def.NowDo()
	do.Operations.Args, do.Rm, do.RmD = []string{"keys"}, true, true
	if err := services.KillService(do); err != nil {
		t.Fatalf("killing keys service failed: %v", err)
	}

	do = def.NowDo()
	do.Name, do.Rm, do.RmD = chain, true, true
	if err := KillChain(do); err != nil {
		t.Fatalf("killing chain failed: %v", err)
	}
}

func exec(t *testing.T, chain string, args []string) string {
	do := def.NowDo()
	do.Name = chain
	do.Operations.Args = args
	buf, err := ExecChain(do)
	if err != nil {
		log.Error(buf)
		t.Fatalf("expected chain to execute, got %v", err)
	}

	return buf.String()
}

func mockChainDefinitionFile(name string) error {
	definition := loaders.MockChainDefinition(name, name, false)

	return WriteChainDefinitionFile(definition, filepath.Join(erisDir, "chains", name))
}
