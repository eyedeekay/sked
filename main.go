package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/eyedeekay/sked/work"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("pass a config file like `example.json`")
		example()
		panic("configerr")
	}
	workFile, runner, err := work.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		runner.Close()
		os.Exit(1)
	}()
	defer runner.Close()
	fmt.Println(workFile)
	fmt.Println(runner.RunDir)
	runner.Run()
}

func example() error {
	command := work.Command{
		Command: "./mirror.sh",
		Args:    []string{},
	}
	command2 := work.Command{
		Command: "./pregenerate.sh",
		Args:    []string{},
	}
	runner := &work.Runner{
		RunDir:          "~/sked-app-dir/",
		Commands:        []work.Command{command, command2},
		IntervalMinutes: 15,
	}
	bytes, err := json.MarshalIndent(runner, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("example.json", bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}
