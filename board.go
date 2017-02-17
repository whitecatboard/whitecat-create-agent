/*
 * Whitecat Blocky Environment, board abstraction
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
	"bytes"
	"encoding/base64"
	"github.com/mikepb/go-serial"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"time"
)

type Board struct {
	// Serial port
	port *serial.Port

	// Board information
	info string

	// RXQueue
	RXQueue chan byte

	// Chunk size for send / receive files to / from board
	chunkSize int

	// If true disables notify board's boot events
	disableInspectorBootNotify bool

	consoleOut bool
}

// Inspects the serial data received for a board in order to find special
// special events, such as reset, core dumps, exceptions, etc ...
//
// Once inspected all bytes are send to RXQueue channel
func (board *Board) inspector() {
	var re *regexp.Regexp

	buffer := make([]byte, 1)

	line := ""
	for {
		if n, err := board.port.Read(buffer); err != nil {
			break
		} else {
			if n > 0 {
				if buffer[0] == '\n' {
					log.Println(line)

					if !board.disableInspectorBootNotify {
						re = regexp.MustCompile(`^rst:.*\(POWERON_RESET\),boot:.*(.*)$`)
						if re.MatchString(line) {
							notify("boardPowerOnReset", "")
						}

						re = regexp.MustCompile(`^rst:.*(SW_CPU_RESET),boot:.*(.*)$`)
						if re.MatchString(line) {
							notify("boardSoftwareReset", "")
						}

						re = regexp.MustCompile(`^rst:.*(DEEPSLEEP_RESET),boot.*(.*)$`)
						if re.MatchString(line) {
							notify("boardDeepSleepReset", "")
						}

						re = regexp.MustCompile(`\<blockStart,(.*)\>`)
						if re.MatchString(line) {
							parts := re.FindStringSubmatch(line)
							info := "\"block\": \"" + base64.StdEncoding.EncodeToString([]byte(parts[1])) + "\""
							notify("blockStart", info)
						}

						re = regexp.MustCompile(`\<blockEnd,(.*)\>`)
						if re.MatchString(line) {
							parts := re.FindStringSubmatch(line)
							info := "\"block\": \"" + base64.StdEncoding.EncodeToString([]byte(parts[1])) + "\""
							notify("blockEnd", info)
						}

						re = regexp.MustCompile(`\<blockError,(.*),(.*)\>`)
						if re.MatchString(line) {
							parts := re.FindStringSubmatch(line)
							info := "\"block\": \"" + base64.StdEncoding.EncodeToString([]byte(parts[1])) + "\", " +
								"\"error\": \"" + base64.StdEncoding.EncodeToString([]byte(parts[2])) + "\""

							notify("blockError", info)
						}
					}

					re = regexp.MustCompile(`^([a-zA-Z]*):(\d*)\:\s(\d*)\:(.*)$`)
					if re.MatchString(line) {
						parts := re.FindStringSubmatch(line)

						info := "\"where\": \"" + parts[1] + "\", " +
							"\"line\": \"" + parts[2] + "\", " +
							"\"exception\": \"" + parts[3] + "\", " +
							"\"message\": \"" + parts[4] + "\""

						notify("boardRuntimeError", info)
					} else {
						re = regexp.MustCompile(`^([a-zA-Z]*)\:(\d*)\:\s*(.*)$`)
						if re.MatchString(line) {
							parts := re.FindStringSubmatch(line)

							info := "\"where\": \"" + parts[1] + "\", " +
								"\"line\": \"" + parts[2] + "\", " +
								"\"exception\": \"0\", " +
								"\"message\": \"" + parts[3] + "\""

							notify("boardRuntimeError", info)
						}
					}

					line = ""
				} else {
					if buffer[0] != '\r' {
						line = line + string(buffer[0])
					}
				}

				//if board.consoleOut {
				//	notify("boardConsoleOut", base64.StdEncoding.EncodeToString(buffer))
				//}

				board.RXQueue <- buffer[0]
			}
		}
	}
}

func (board *Board) attach(info *serial.Info) bool {
	// Configure options or serial port connection
	options := serial.RawOptions
	options.BitRate = 115200
	options.Mode = serial.MODE_READ_WRITE
	options.DTR = serial.DTR_OFF
	options.RTS = serial.RTS_OFF

	// Open port
	port, openErr := options.Open(info.Name())
	if openErr != nil {
		return false
	}

	// Create board struct
	board.port = port
	board.RXQueue = make(chan byte, 10*1024)
	board.chunkSize = 255
	board.disableInspectorBootNotify = false
	board.consoleOut = false

	go board.inspector()

	// Reset the board
	if board.reset(true) {
		connectedBoard = board

		notify("boardAttached", "")

		return true
	}

	return false
}

func (board *Board) detach() {
	// Close board
	if board != nil {
		board.consume()
		board.port.Close()
	}

	connectedBoard = nil
}

/*
 * Serial port primitives
 */

// Read one byte from RXQueue
func (board *Board) read() byte {
	return <-board.RXQueue
}

// Read one line from RXQueue
func (board *Board) readLine() string {
	var buffer bytes.Buffer
	var b byte

	for {
		b = board.read()
		if b == '\n' {
			return buffer.String()
		} else {
			if b != '\r' {
				buffer.WriteString(string(rune(b)))
			}
		}
	}

	return ""
}

func (board *Board) consume() {
	time.Sleep(time.Millisecond * 100)

	for len(board.RXQueue) > 0 {
		board.read()
	}
}

// Wait until board is ready
func (board *Board) waitForReady() bool {
	//	timeout := 0
	booting := false
	whitecat := false
	line := ""

	for {
		line = board.readLine()

		if !booting {
			booting = regexp.MustCompile(`^rst:.*\(POWERON_RESET\),boot:.*(.*)$`).MatchString(line)
		} else {
			if !whitecat {
				whitecat = regexp.MustCompile(`whitecatboard\.org`).MatchString(line)
				if whitecat {
					// Send Ctrl-D
					board.port.Write([]byte{4})
				}
			} else {
				if regexp.MustCompile(`^Lua RTOS-boot-scripts-aborted-ESP32$`).MatchString(line) {
					return true
				}
			}
		}
	}
}

// Test if line corresponds to Lua RTOS prompt
func isPrompt(line string) bool {
	return regexp.MustCompile("^/.*>.*$").MatchString(line)
}

func (board *Board) getInfo() string {
	info := board.sendCommand("dofile(\"/_info.lua\")")

	info = strings.Replace(info, ",}", "}", -1)
	info = strings.Replace(info, ",]", "]", -1)

	return info
}

// Send a command to the board
func (board *Board) sendCommand(command string) string {
	var response string = ""

	// Send command. We must append the \r\n chars at the end
	board.port.Write([]byte(command + "\r\n"))

	// Read response, that it must be the send command.
	line := board.readLine()
	if line == command {
		// Read until prompt
		for {
			line = board.readLine()

			if isPrompt(line) {
				return response
			} else {
				if response != "" {
					response = response + "\r\n"
				}
				response = response + line
			}
		}
	} else {
		return ""
	}

	return ""
}

func (board *Board) reset(prerequisites bool) bool {
	board.consume()

	// Reset board
	options := serial.RawOptions
	options.BitRate = 115200
	options.Mode = serial.MODE_READ_WRITE

	options.RTS = serial.RTS_OFF
	board.port.Apply(&options)

	time.Sleep(time.Millisecond * 10)

	options.RTS = serial.RTS_ON
	board.port.Apply(&options)

	time.Sleep(time.Millisecond * 10)

	options.RTS = serial.RTS_OFF
	board.port.Apply(&options)

	if !board.waitForReady() {
		log.Println("board timeout ...")

		return false
	}

	board.consume()

	log.Println("board is ready ...")

	if prerequisites {
		buffer, err := ioutil.ReadFile("./boards/lua/board-info.lua")
		if err == nil {
			board.writeFile("/_info.lua", buffer)
		}

		files, err := ioutil.ReadDir("./boards/lua/lib")
		if err == nil {
			for _, finfo := range files {
				if regexp.MustCompile(`.*\.lua`).MatchString(finfo.Name()) {
					file, _ := ioutil.ReadFile("./boards/lua/lib/" + finfo.Name())
					board.consume()
					board.writeFile("/lib/lua/"+finfo.Name(), file)
				}
			}
		}
	}

	// Get board info
	info := board.getInfo()

	log.Println("board info: ", info)

	board.info = info
	return true
}

func (board *Board) getDirContent(path string) string {
	var content = ""

	response := board.sendCommand("os.ls(\"" + path + "\")")
	for _, line := range strings.Split(response, "\n") {
		element := strings.Split(strings.Replace(line, "\r", "", -1), "\t")

		if len(element) == 4 {
			if content != "" {
				content = content + ","
			}

			content = content + "{" +
				"\"type\": \"" + element[0] + "\"," +
				"\"size\": \"" + element[1] + "\"," +
				"\"date\": \"" + element[2] + "\"," +
				"\"name\": \"" + element[3] + "\"" +
				"}"
		}
	}

	return "[" + content + "]"
}

func (board *Board) writeFile(path string, buffer []byte) {
	writeCommand := "io.receive(\"" + path + "\")"

	outLen := 0
	outIndex := 0

	// Send command and test for echo
	board.port.Write([]byte(writeCommand + "\r"))
	if board.readLine() == writeCommand {
		for {
			// Wait for chunk
			if board.readLine() == "C" {
				// Get chunk length
				if outIndex < len(buffer) {
					if outIndex+board.chunkSize-1 < len(buffer) {
						outLen = board.chunkSize
					} else {
						outLen = len(buffer) - outIndex
					}
				} else {
					outLen = 0
				}

				// Send chunk length
				board.port.Write([]byte{byte(outLen)})

				if outLen > 0 {
					// Send chunk
					board.port.Write(buffer[outIndex : outIndex+outLen])
				} else {
					break
				}

				outIndex = outIndex + outLen
			}
		}

		board.consume()
	}
}

func (board *Board) readFile(path string) []byte {
	var buffer bytes.Buffer
	var inLen byte

	// Command for read file
	readCommand := "io.send(\"" + path + "\")"

	// Send command and test for echo
	board.port.Write([]byte(readCommand + "\r"))
	if board.readLine() == readCommand {
		for {
			// Wait for chunk
			board.port.Write([]byte("C\n"))

			// Read chunk size
			inLen = board.read()

			// Read chunk
			if inLen > 0 {
				for inLen > 0 {
					buffer.WriteByte(board.read())

					inLen = inLen - 1
				}
			} else {
				// No more data
				break
			}
		}

		board.consume()

		return buffer.Bytes()
	}

	return nil
}

func (board *Board) runProgram(path string, code []byte) {
	board.disableInspectorBootNotify = true

	// Reset board
	if board.reset(false) {
		board.disableInspectorBootNotify = false

		// First update autorun.lua, which run the target file
		board.writeFile("/autorun.lua", []byte("dofile(\""+path+"\")\r\n"))

		// Now write code to target file
		board.writeFile(path, code)

		// Run the target file
		board.port.Write([]byte("require(\"block\");wcBlock.delevepMode=true;dofile(\"" + path + "\")\r"))

		board.consume()
	}
}

func (board *Board) runCommand(code []byte) string {
	result := board.sendCommand(string(code))
	board.consume()

	return result
}
