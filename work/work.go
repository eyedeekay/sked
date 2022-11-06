package work

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Command struct {
	Command string
	Args    []string
}

type Runner struct {
	RunDir          string
	Commands        []Command
	IntervalMinutes int
	workFile        string
}

func (runner *Runner) Open(configFile string) (workFile string, err error) {
	runner.workFile, *runner, err = Open(configFile)
	return runner.workFile, err
}

func (runner *Runner) createWorkFile() error {
	return ioutil.WriteFile(runner.workFile, []byte(""), 0644)
}

func (runner *Runner) removeWorkFile() error {
	return os.RemoveAll(runner.workFile)
}

func (runner *Runner) Run() error {
	defer runner.Close()
	os.Chdir(runner.RunDir)
	for {
		if err := runner.createWorkFile(); err != nil {
			return err
		}
		for _, run := range runner.Commands {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			run.Command = strings.Replace(run.Command, ".", wd, 1)
			cmd := exec.Command(run.Command, run.Args...)
			stderr, err := cmd.StderrPipe()
			if err != nil {
				log.Fatalf("could not get stderr pipe: %v", err)
			}
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				log.Fatalf("could not get stdout pipe: %v", err)
			}
			go func() {
				merged := io.MultiReader(stderr, stdout)
				scanner := bufio.NewScanner(merged)
				for scanner.Scan() {
					msg := scanner.Text()
					fmt.Printf("msg: %s\n", msg)
				}
			}()
			fmt.Println("Running:", cmd.String())
			fmt.Println("Within:", wd)
			err = cmd.Run()
			if err != nil {
				return err
			}
			cmd.Stderr = cmd.Stdout
			time.Sleep(time.Second)
		}
		tval := (time.Minute * time.Duration(runner.IntervalMinutes)) - (time.Second * time.Duration(len(runner.Commands)))
		fmt.Println("Sleeping for:", tval, "minutes")
		time.Sleep(tval)
		if _, err := os.Stat(runner.workFile); errors.Is(err, os.ErrNotExist) {
			return nil
		}
		runner.removeWorkFile()
	}
}

func (runner *Runner) Close() error {
	return Close(runner.workFile)
}

func Open(configFile string) (workFile string, runner Runner, err error) {
	var content []byte
	if content, err = ioutil.ReadFile(configFile); err != nil {
		return
	} else {
		var userHome string
		userHome, err = os.UserHomeDir()
		if err != nil {
			return
		}
		content = []byte(strings.ReplaceAll(string(content), "~", userHome))
		if err = json.Unmarshal(content, &runner); err != nil {
			return
		}
		workFile = filepath.Join(runner.RunDir, "command-is-running")
		runner.workFile = workFile
		return
	}
}

func Close(workFile string) error {
	if _, err := os.Stat(workFile); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err := os.RemoveAll(workFile); err != nil {
		return err
	}
	return nil
}
