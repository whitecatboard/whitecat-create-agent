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
	"fmt"
	"github.com/kardianos/osext"
	"github.com/mikepb/go-serial"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"runtime"
	"strconv"
	"time"
)

var Version string = "1.3"
var Options []string

var Upgrading bool = false
var StopMonitor bool = false

var AppFolder = "/"
var AppDataFolder string = "/"
var AppDataTmpFolder string = "/tmp"
var AppFileName = ""

// Connected board
var connectedBoard *Board = nil

// Monitor serial ports and search for a board compatible with Lua RTOS.
// If a board is found, monitors that port continues open over time.
func monitorSerialPorts(devices []deviceDef) {
	previousPorts := 0
	
	// This variable computes the elapsed time monitoring serial ports without exit
	elapsed := 0
		
	log.Println("start monitoring serial ports ...")

	notify("boardUpdate", "Scanning boards")
	
	for {
		if Upgrading {
			time.Sleep(time.Millisecond * 500)
			continue
		}

		if StopMonitor {
			log.Println("Stop monitoring serial ports ...")
			break
		}

		// If a board is connected ...
		if connectedBoard != nil {
			// Test that port continues open
			_, err := connectedBoard.port.InputWaiting()
			if err != nil {
				notify("boardDetached", "")

				connectedBoard.detach()
				
				// Port is not open, waiting for a board
				connectedBoard = nil				
			} else {
				// Port is open, continue
				elapsed = 0
				time.Sleep(time.Millisecond * 500)
				continue
			}
		}

		// Enumerate all serial ports
		ports, err := serial.ListPorts()
		if err != nil {
			previousPorts = 0
			continue
		}
		
		// Update port count, if there is a changed there are something new connected, so
		// inform the IDE that the Agent is scanning
		if (len(ports) != previousPorts) {
			notify("boardUpdate", "Scanning boards")
		}
		previousPorts = len(ports)
		
		
		// Search a serial port that matches with one of the supported adapters
		for _, info := range ports {
			// Read VID/PID
			vendorId, productId, err := info.USBVIDPID()
			if err != nil {
				continue
			}

			// We need a VID / PID
			if vendorId != 0 && productId != 0 {
				vendorId := "0x" + strconv.FormatInt(int64(vendorId), 16)
				productId := "0x" + strconv.FormatInt(int64(productId), 16)

				log.Printf("found adapter, VID %s:%s", vendorId, productId)

				// Search a VID/PIN into requested devices
				for _, device := range devices {
					if device.VendorId == vendorId && device.ProductId == productId {
						// This adapter matches

						log.Printf("check adapter, VID %s:%s", device.VendorId, device.ProductId)

						// Create a candidate board
						var candidate Board

						// Attach candidate
						if candidate.attach(info) {
							break
						}
					}
				}
			}
		}

		time.Sleep(time.Millisecond * 10)

		if connectedBoard == nil {
			elapsed = elapsed + 10
			if elapsed > 5000 {
				// No board found in 5 seconds
				notify("boardUpdate", "No board attached")

				elapsed = 0
			}
		}
	}
}

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
