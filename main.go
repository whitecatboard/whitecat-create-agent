/*
 * Whitecat Blocky Environment, agent main program
 *
 * Copyright (C) 2015 - 2016
 * IBEROXARXA SERVICIOS INTEGRALES, S.L.
 *
 * Author: Jaume Olivé (jolive@iberoxarxa.com / jolive@whitecatboard.org)
 *
 * All rights reserved.
 *
 * Permission to use, copy, modify, and distribute this software
 * and its documentation for any purpose and without fee is hereby
 * granted, provided that the above copyright notice appear in all
 * copies and that both that the copyright notice and this
 * permission notice and warranty disclaimer appear in supporting
 * documentation, and that the name of the author not be used in
 * advertising or publicity pertaining to distribution of the
 * software without specific, written prior permission.
 *
 * The author disclaim all warranties with regard to this
 * software, including all implied warranties of merchantability
 * and fitness.  In no event shall the author be liable for any
 * special, indirect or consequential damages or any damages
 * whatsoever resulting from loss of use, data or profits, whether
 * in an action of contract, negligence or other tortious action,
 * arising out of or in connection with the use or performance of
 * this software.
 */

package main

import (
	"bytes"
	"fmt"
	"github.com/kardianos/osext"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"runtime"
	"strings"
)

var Version string = "1.3"
var Options []string

var AppFolder = "/"
var AppDataFolder string = "/"
var AppDataTmpFolder string = "/tmp"
var AppFileName = ""
var PythonPath = ""

func usage() {
	fmt.Println("wccagent: usage: wccagent [-lf | -lc | -ui | -v]")
	fmt.Println("")
	fmt.Println(" -b : run in background (only windows)")
	fmt.Println(" -lf: log to file")
	fmt.Println(" -lc: log to console")
	fmt.Println(" -ui: enable the user interface")
	fmt.Println(" -v : show version")
}

func restart() {
	if runtime.GOOS == "darwin" {
		os.Exit(1)
	} else {
		cmd := exec.Command(AppFileName, "-ui")
		cmd.Start()
		os.Exit(0)
	}
}

func start(ui bool, background bool) {
	if ui {
		if background {
			restart()
		} else {
			setupSysTray()
		}
	} else {
		exitChan := make(chan int)

		go webSocketStart(exitChan)
		<-exitChan
	}
}

func main() {
	includeInRespawn := false
	withLogFile := false
	withLogConsole := false
	withUI := false
	withBackground := false
	ok := true
	i := 0

	// Get arguments and process arguments
	for _, arg := range os.Args {
		includeInRespawn = true

		switch arg {
		case "-b":
			if runtime.GOOS == "windows" {
				withBackground = true
			} else {
				ok = false
			}
		case "-lf":
			withLogFile = true
		case "-lc":
			withLogConsole = true
		case "-ui":
			withUI = true
		case "-v":
			includeInRespawn = false
			fmt.Println(Version)
			os.Exit(0)
		default:
			if i > 0 {
				ok = false
			}
		}

		if includeInRespawn && (i > 0) {
			Options = append(Options, arg)
		}

		i = i + 1
	}

	if !ok {
		usage()
		os.Exit(1)
	}

	// Get home directory, create the user data folder, and needed folders
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	if runtime.GOOS == "darwin" {
		AppDataFolder = path.Join(usr.HomeDir, ".wccagent")
	} else if runtime.GOOS == "windows" {
		AppDataFolder = path.Join(usr.HomeDir, "AppData", "The Whitecat Create Agent")
	} else if runtime.GOOS == "linux" {
		AppDataFolder = path.Join(usr.HomeDir, ".whitecat-create-agent")
	}

	AppDataTmpFolder = path.Join(AppDataFolder, "tmp")

	_ = os.Mkdir(AppDataFolder, 0755)
	_ = os.Mkdir(AppDataTmpFolder, 0755)

	// Get where program is executed
	execFolder, err := osext.ExecutableFolder()
	if err != nil {
		panic(err)
	}

	AppFolder = execFolder
	AppFileName, _ = osext.Executable()

	// Get python path
	var out bytes.Buffer
	var stderr bytes.Buffer
	var cmd *exec.Cmd

	if runtime.GOOS == "darwin" {
		cmd = exec.Command("/usr/bin/whereis", "python")
	} else if runtime.GOOS == "windows" {
		cmd = exec.Command("where", "python.exe")
	} else if runtime.GOOS == "linux" {
		cmd = exec.Command("/usr/bin/whereis", "-b", "python")
	}

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	PythonPath = out.String()

	PythonPath = strings.Replace(PythonPath, "python:", "", -1)
	PythonPath = strings.Replace(PythonPath, "\r", "", -1)

	PythonPath = strings.Split(PythonPath, "\n")[0]

	// Set log options
	if withLogConsole {
		// User wants log to console
	} else if withLogFile {
		// User wants log to file
		f, _ := os.OpenFile(path.Join(AppDataFolder, "log.txt"), os.O_RDWR|os.O_CREATE, 0755)
		log.SetOutput(f)
		defer f.Close()
	} else {
		// User does not want log
		log.SetOutput(ioutil.Discard)
	}

	start(withUI, withBackground)
}
