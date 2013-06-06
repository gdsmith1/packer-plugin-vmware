package vmware

import (
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"time"
)

const BuilderId = "mitchellh.vmware"

type Builder struct {
	config config
	runner multistep.Runner
}

type config struct {
	DiskName       string   `mapstructure:"vmdk_name"`
	ISOUrl         string   `mapstructure:"iso_url"`
	VMName         string   `mapstructure:"vm_name"`
	OutputDir      string   `mapstructure:"output_directory"`
	HTTPDir        string   `mapstructure:"http_directory"`
	BootCommand    []string `mapstructure:"boot_command"`
	BootWait       time.Duration
	SSHUser        string `mapstructure:"ssh_user"`
	SSHPassword    string `mapstructure:"ssh_password"`
	SSHWaitTimeout time.Duration

	RawBootWait       string `mapstructure:"boot_wait"`
	RawSSHWaitTimeout string `mapstructure:"ssh_wait_timeout"`
}

func (b *Builder) Prepare(raw interface{}) (err error) {
	err = mapstructure.Decode(raw, &b.config)
	if err != nil {
		return
	}

	if b.config.DiskName == "" {
		b.config.DiskName = "disk"
	}

	if b.config.VMName == "" {
		b.config.VMName = "packer"
	}

	if b.config.OutputDir == "" {
		b.config.OutputDir = "vmware"
	}

	if b.config.RawBootWait != "" {
		b.config.BootWait, err = time.ParseDuration(b.config.RawBootWait)
		if err != nil {
			return
		}
	}

	if b.config.RawSSHWaitTimeout == "" {
		b.config.RawSSHWaitTimeout = "20m"
	}

	b.config.SSHWaitTimeout, err = time.ParseDuration(b.config.RawSSHWaitTimeout)
	if err != nil {
		return
	}

	return nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook) packer.Artifact {
	steps := []multistep.Step{
		&stepPrepareOutputDir{},
		&stepCreateDisk{},
		&stepCreateVMX{},
		&stepHTTPServer{},
		&stepRun{},
		&stepTypeBootCommand{},
		&stepWaitForSSH{},
		&stepProvision{},
	}

	// Setup the state bag
	state := make(map[string]interface{})
	state["config"] = &b.config
	state["hook"] = hook
	state["ui"] = ui

	// Run!
	b.runner = &multistep.BasicRunner{Steps: steps}
	b.runner.Run(state)

	return nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
