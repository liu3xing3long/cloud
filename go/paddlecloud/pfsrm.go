package paddlecloud

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/PaddlePaddle/cloud/go/filemanager/pfsmod"
	log "github.com/golang/glog"
	"github.com/google/subcommands"
)

type RmCommand struct {
	cmd pfsmod.RmCmd
}

func (*RmCommand) Name() string     { return "rm" }
func (*RmCommand) Synopsis() string { return "rm files on PaddlePaddle Cloud" }
func (*RmCommand) Usage() string {
	return `rm -r <pfspath>:
	rm files on PaddlePaddleCloud
	Options:
`
}

func (p *RmCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.cmd.R, "r", false, "rm files recursively")
}

func formatRmPrint(results []pfsmod.RmResult, err error) {
	for _, result := range results {
		fmt.Printf("rm %s\n", result.Path)
	}

	if err != nil {
		fmt.Println("\t" + err.Error())
	}

	return
}

func RemoteRm(s *PfsSubmitter, cmd *pfsmod.RmCmd) ([]pfsmod.RmResult, error) {
	body, err := s.PostFiles(cmd)
	if err != nil {
		return nil, err

	}

	log.V(3).Info(string(body[:]))

	resp := pfsmod.RmResponse{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return resp.Results, err
	}

	log.V(1).Infof("%#v\n", resp)

	if len(resp.Err) == 0 {
		return resp.Results, nil
	}

	return resp.Results, errors.New(resp.Err)
}

func remoteRm(s *PfsSubmitter, cmd *pfsmod.RmCmd) error {
	for _, arg := range cmd.Args {
		subcmd := pfsmod.NewRmCmd(
			cmd.R,
			arg,
		)

		fmt.Printf("rm %s\n", arg)
		result, err := RemoteRm(s, subcmd)
		formatRmPrint(result, err)
	}
	return nil

}

func (p *RmCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		f.Usage()
		return subcommands.ExitFailure
	}

	cmd, err := pfsmod.NewRmCmdFromFlag(f)
	if err != nil {
		return subcommands.ExitFailure
	}
	log.V(1).Infof("%#v\n", cmd)

	s := NewPfsCmdSubmitter(UserHomeDir() + "/.paddle/config")
	if err := remoteRm(s, cmd); err != nil {
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
