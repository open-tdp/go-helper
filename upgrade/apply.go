package upgrade

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/minio/selfupdate"
	"github.com/open-tdp/go-helper/logman"
)

func Apply(rq *RequesParam) error {

	logger := logman.Named("upgrade")

	logger.Info(
		"checking update",
		"version", rq.Version,
		"url", rq.Server,
	)

	resp, err := CheckVersion(rq)
	if err != nil {
		logger.Error("check update failed", "error", err)
		return err
	}

	if !strings.HasPrefix(resp.Package, "https://") {
		logger.Info("no need to update", "resp", resp)
		return ErrNoUpdate
	}

	updater, err := Downloader(resp)
	if err != nil {
		logger.Error("prepare updater failed", "error", err)
		return err
	}

	defer updater.Close()

	err = selfupdate.Apply(updater, selfupdate.Options{})
	if err != nil {
		logger.Error("apply update failed", "error", err)
		if selfupdate.RollbackError(err) != nil {
			logger.Error("failed to rollback from bad update")
		}
		return err
	}

	return nil

}

func Restart() error {

	self, err := os.Executable()
	if err != nil {
		return err
	}

	args, env := os.Args, os.Environ()

	// Windows does not support exec syscall
	if runtime.GOOS == "windows" {
		cmd := exec.Command(self, args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = env
		err := cmd.Start()
		if err == nil {
			os.Exit(0)
		}
		return err
	}

	// Other OS
	return syscall.Exec(self, args, env)

}
