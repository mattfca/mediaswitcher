package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// Config is just the stuff we need to run child miners
type Config struct {
	Devices          string
	FallbackThreads  string
	X16RUser         string
	X16RWorker       string
	X16RPassword     string
	EquihashUser     string
	EquihashWorker   string
	EquihashPassword string
	EquihashServer   string
	EquihashPort     string
}

var conf Config
var EquihashCMD *exec.Cmd
var CMD *exec.Cmd

func main() {
	flag.StringVar(&conf.Devices, "d", "", "Devices (default auto)")
	flag.StringVar(&conf.FallbackThreads, "f", "8", "Fallback threads (default 1)")
	flag.StringVar(&conf.X16RUser, "u1", "mediafraze", "X16R Pool user or address (default me ;))")
	flag.StringVar(&conf.X16RWorker, "w1", "mediaswitcher", "X16R Worker name (dafault mediaswitcher)")
	flag.StringVar(&conf.X16RPassword, "p1", "x", "X16R Worker password (default x)")

	flag.StringVar(&conf.EquihashServer, "es", "us-east.equihash-hub.miningpoolhub.com", "Equihash server (default mph)")
	flag.StringVar(&conf.EquihashPort, "ep", "17023", "Equihash port (default auto algo)")
	flag.StringVar(&conf.EquihashUser, "eu", "mediafraze", "Equihash Pool user or address (default me ;))")
	flag.StringVar(&conf.EquihashWorker, "ew", "mediaswitcher", "Equihash Worker name (dafault mediaswitcher)")
	flag.StringVar(&conf.EquihashPassword, "epw", "x", "Equihash Worker password (default x)")
	flag.Parse()

	fmt.Printf("%#v\n", conf)

	runX16R()
}

func runX16R() {
	args := fmt.Sprintf("-a x16r -o stratum+tcp://rvn.suprnova.cc:6667 -u %s.%s -p %s -d %s  --num-fallback-threads=%s", conf.X16RUser, conf.X16RWorker, conf.X16RPassword, conf.Devices, conf.FallbackThreads)
	fmt.Println("Running: miners/ccminer-x64.exe %s", args)

	CMD = exec.Command("miners/ccminer-x64.exe", strings.Split(args, " ")...)

	/*
	* X16R
	 */

	stdout, err := CMD.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(stdout)
	cpu := false
	gpu := false

	go func() {
		for scanner.Scan() {
			ln := scanner.Text()
			cpuPrev := cpu
			gpuPrev := gpu

			fmt.Println(ln)

			if strings.Contains(ln, "Partial GPU job") {
				cpu = true
				gpu = false
			}

			if strings.Contains(ln, "100% GPU job") {
				cpu = false
				gpu = true
			}

			if cpu && !cpuPrev {
				fmt.Println("CPU Mining Detected")
				go runEquihash()
			}

			if gpu && !gpuPrev {
				fmt.Println("GPU Mining Detected")

				if EquihashCMD != nil {
					fmt.Println("Stopping Equihash")
					if err := EquihashCMD.Process.Kill(); err != nil {
						fmt.Println("Failed to stop dstm")
					}
				}

			}
		}
	}()

	CMD.Start()
	CMD.Wait()
}

func runEquihash() {
	fmt.Println("Starting Equihash")

	devices := strings.Replace(conf.Devices, ",", " ", -1)

	equihashArgs := fmt.Sprintf("--d %s --server %s --port %s --user %s.%s --pass %s", devices, conf.EquihashServer, conf.EquihashPort, conf.EquihashUser, conf.EquihashWorker, conf.EquihashPassword)
	EquihashCMD = exec.Command("miners/zm_0.5.7_win/zm.exe", strings.Split(equihashArgs, " ")...)
	fmt.Println("miners/zm_0.5.7_win/zm.exe %s", equihashArgs)
	/*
	* Equihash
	 */

	stdoutEquihash, err := EquihashCMD.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	scannerEquihash := bufio.NewScanner(stdoutEquihash)

	go func() {
		for scannerEquihash.Scan() {
			ln := scannerEquihash.Text()
			fmt.Println(ln)
		}
	}()

	EquihashCMD.Run()
	EquihashCMD.Wait()
}
