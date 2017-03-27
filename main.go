/*
 * Whitecat Blocky Environment, agent main program
 *
 * Copyright (C) 2015 - 2016
 * IBEROXARXA SERVICIOS INTEGRALES, S.L. & CSS IBÉRICA, S.L.
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
	"github.com/mikepb/go-serial"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

var Upgrading bool = false
var StopMonitor bool = false

// Connected board
var connectedBoard *Board = nil

// Monitor serial ports and search for a board compatible with Lua RTOS.
// If a board is found, monitors that port continues open over time.
func monitorSerialPorts(devices []deviceDef) {
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

				// Port is not open, waiting for a board
				connectedBoard = nil
			} else {
				// Port is open, continue
				time.Sleep(time.Millisecond * 10)
				continue
			}
		}

		// Enumerate all serial ports
		ports, err := serial.ListPorts()
		if err != nil {
			continue
		}

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
	}
}

func main() {
	withLog := false

	// Get arguments and process arguments
	for _, arg := range os.Args {
		switch arg {
		case "-l":
			withLog = true
		}
	}

	if !withLog {
		log.SetOutput(ioutil.Discard)
		setupSysTray()
	} else {
		exitChan := make(chan int)

		go webSocketStart(exitChan)
		<-exitChan
	}
}
