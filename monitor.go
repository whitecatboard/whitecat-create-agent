/*
 * Whitecat Blocky Environment, serial port monitor
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
	"github.com/mikepb/go-serial"
	"log"
	"strconv"
	"time"
)

// Connected board
var connectedBoard *Board = nil

// This variable computes the elapsed time monitoring serial ports without success
var elapsed int = 0

func tryLater() {
	time.Sleep(time.Millisecond * 10)

	if connectedBoard == nil {
		elapsed = elapsed + 10
		if elapsed > 5000 {
			// No board found in the last 5 seconds
			notify("boardUpdate", "No board attached")

			elapsed = 0
		}
	}
}

// Monitor serial ports and search for a Lua RTOS device.
// If a Lua RTOS device is found monitor the serial port.
func monitor() {
	defer func() {
		log.Println("stop monitor ...")

		if err := recover(); err != nil {
			time.Sleep(time.Millisecond * 1000)
			go monitor()
		}
	}()

	log.Println("start monitor ...")

	// Notify IDE that monitor is searching for a board
	notify("boardUpdate", "Scanning boards")

	for {
		select {
		case <-IdeDetach:
			return
		default:
			if Upgrading {
				time.Sleep(time.Millisecond * 100)
				continue
			}

			// If a board is connected thest that is still connected
			if connectedBoard != nil {
				_, err := connectedBoard.port.InputWaiting()
				if err != nil {
					// Board is not connected, inform the IDE
					connectedBoard.detach()

					notify("boardDetached", "")
					panic(err)
				} else {
					// Board is connected, check again later
					tryLater()
					continue
				}
			}

			// In this point any board is connected, search for a board ...

			// Enumerate all serial ports
			ports, err := serial.ListPorts()
			if err != nil {
				tryLater()
				continue
			}

			// Search a serial port that matches with one of the supported adapters
			for _, info := range ports {
				// Read VID/PID
				vendorId, productId, err := info.USBVIDPID()
				if err != nil {
					time.Sleep(time.Millisecond * 100)
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
							candidate.attach(info)

							if connectedBoard != nil {
								break
							}
						}
					}

					if connectedBoard != nil {
						break
					}
				}
			}

			if connectedBoard == nil {
				tryLater()
			}
		}
	}
}
