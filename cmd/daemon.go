package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/aaronireland/state-server/pkg/server"
)

var waitForSeconds = 15

func daemon(cmd string, action server.Action, args ...string) error {
	procs, err := findProcess(cmd)
	if err != nil {
		return fmt.Errorf("failed to check currently running servers: %w", err)
	}

	if action == server.Shutdown && len(procs) == 0 {
		return nil
	}

	for _, proc := range procs {
		if err := terminateProcess(proc); err != nil {
			return fmt.Errorf("failed to terminate running server (pid: %d): %w", proc.Pid, err)
		} else {
			fmt.Fprintf(os.Stdout, "pid %d terminated...\n", proc.Pid)
		}
	}

	if action == server.Start {
		proc, err := startProcess(cmd, args...)
		if err != nil {
			procName := "server process"
			if proc != nil {
				procName = fmt.Sprintf("pid: %d", proc.Pid)
			}
			return fmt.Errorf("failed to start %s -> %s", procName, err)
		} else {
			fmt.Fprintf(os.Stdout, "pid %d started...\n", proc.Pid)
		}
	}

	fmt.Printf("State Server %s finished.\n", action)
	return nil
}

func findProcess(cmd string) (procs []*os.Process, err error) {
	bin := filepath.Base(cmd)
	start, stop := cmd+" start", cmd+" stop"

	psOutput, err := exec.Command("ps", "-e").Output()
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(bytes.NewReader(psOutput))
	for scanner.Scan() {
		ps := scanner.Text()
		ignore := strings.Contains(ps, start) || strings.Contains(ps, stop)
		if strings.Contains(ps, bin) && !ignore {
			psPid := strings.FieldsFunc(ps, unicode.IsSpace)[0]
			pid, err := strconv.Atoi(strings.TrimSpace(psPid))
			if err != nil {
				return procs, err
			}

			if proc, err := getProcess(pid); err != nil {
				log.Println(err.Error())
			} else if proc != nil {
				procs = append(procs, proc)
			}
		}
	}

	return
}

func getProcess(pid int) (proc *os.Process, err error) {
	proc, err = os.FindProcess(pid)
	if err != nil {
		return
	}
	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return
	}

	if errors.Is(err, os.ErrProcessDone) {
		return nil, nil
	}

	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return nil, err
	}
	switch errno {
	case syscall.ESRCH:
		return nil, nil
	case syscall.EPERM:
		return proc, nil
	}

	return nil, err
}

func terminateProcess(proc *os.Process) error {
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	shutdown := make(chan error, 1)
	go waitFor(server.Shutdown, proc.Pid, shutdown)
	return <-shutdown
}

func startProcess(cmd string, args ...string) (proc *os.Process, err error) {
	service := exec.Command(cmd, args...)
	err = service.Start()
	if err != nil {
		return
	}
	proc = service.Process

	started := make(chan error, 1)
	go waitFor(server.Start, proc.Pid, started)
	err = <-started

	return
}

func waitFor(action server.Action, pid int, results chan error) {
	desired := action == server.Start // waitFor IsRunning = true for start, false for shutdown
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	timeout := time.NewTimer(time.Duration(waitForSeconds) * time.Second)
	defer timeout.Stop()

	for {
		select {
		case <-timeout.C:
			results <- fmt.Errorf("timeout waiting for process %s", action)
			return
		case <-ticker.C:
			proc, err := getProcess(pid)
			running := proc != nil && err == nil
			if running == desired {
				results <- nil
				return
			}
		}
	}
}
