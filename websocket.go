/*
 * Whitecat Blocky Environment, Whitecat Agent Websocket Server
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

/*

This is the implementation for the Whitecat Agent websocket server. This server has two main
functions:

* Listen commands sended by the IDE, process this commands, and send back a response to the IDE
* Notify IDE about important things that it must be know

Notifications:

{"notify": "boardAttached", "info": {"modules":[], "maps": []}}
{"notify": "boardDetached", "info": {}}
{"notify": "boardPowerOnReset", "info": {}}
{"notify": "boardSoftwareReset", "info": {}}
{"notify": "boardDeepSleepReset", "info": {}}
{"notify": "boardRuntimeError", "info": {"where": "xx", "line": "xx", "exception": "xx", "message": "xx"}}
{"notify": "boardConsoleOut", "info": {"content": "xxx"}}
{"notify": "boardUptate", "info": {}}
{"notify": "boardUpgraded", "info": {}}

Available commands:

{"command": "attachIde", "arguments": "{}"}
{"command": "detachIde", "arguments": "{}"}

{"command": "boardUpgrade", "arguments": "{}"}
{"command": "boardInfo", "arguments": "{}"}
{"command": "boardReset, "arguments": "{}"}
{"command": "boardStop, "arguments": "{}"}
{"command": "boardGetDirContent", "arguments": {"path": "xxxx"}}
{"command": "boardReadFile", "arguments": {"path": "xxxx"}}
{"command": "boardRunProgram", "arguments": {"path": "xxxx", "code": "xxxx"}}
{"command": "boardRunCommand", "arguments": {"code": "xxxx"}}

*/

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
)

var WS *websocket.Conn = nil

type deviceDef struct {
	VendorId  string
	ProductId string
	Vendor    string
}

var devices []deviceDef

type Command struct {
	Command string
}

type CommandFileSystem struct {
	Command   string
	Arguments struct {
		Path    string
		Content string
	}
}

type CommandRunProgram struct {
	Command   string
	Arguments struct {
		Path string
		Code string
	}
}

type CommandRunCommand struct {
	Command   string
	Arguments struct {
		Path string
		Code string
	}
}

type AttachIdeCommand struct {
	Command   string
	Arguments struct {
		Devices []deviceDef
	}
}

func notify(notification string, data string) {
	var err error
	var msg string
	var info string = "{}"

	// Build info for each notification type
	switch notification {
	case "boardAttached":
		newBuild := "false"
		if connectedBoard.newBuild {
			newBuild = "true"
		}

		info = "{\"info\": " + connectedBoard.info + ", \"newBuild\": " + newBuild + "}"

	case "blockStart":
		info = "{" + data + "}"

	case "blockEnd":
		info = "{" + data + "}"

	case "blockError":
		info = "{" + data + "}"

	case "boardRuntimeError":
		info = "{" + data + "}"

	case "boardGetDirContent":
		info = data

	case "boardReadFile":
		info = "{\"content\": \"" + data + "\"}"

	case "boardConsoleOut":
		info = "{\"content\": \"" + data + "\"}"

	case "boardRunCommand":
		info = "{\"response\": \"" + data + "\"}"

	case "boardUpdate":
		info = "{\"what\": \"" + data + "\"}"

	case "attachIde":
		info = "{\"agent-version\": \"" + Version + "\"}"
	}

	// Build message
	msg = "{\"notify\": \"" + notification + "\", \"info\": " + info + "}"

	// Send message
	if WS != nil {
		if err = websocket.Message.Send(WS, msg); err != nil {
			fmt.Println("Can't send")
		}
		log.Println("notify: ", msg)
	} else {
		log.Println("can't notify: ", msg)
	}
}

func handler(ws *websocket.Conn) {
	var msg string
	var err error
	var command Command

	WS = ws

	log.Println("websocket new connection ...")

	for {
		// Get a new message
		if err = websocket.Message.Receive(ws, &msg); err != nil {
			break
		}

		log.Println("received message: ", msg)

		// Parse command
		json.Unmarshal([]byte(msg), &command)

		switch command.Command {
		case "attachIde":
			StopMonitor = false
			if connectedBoard == nil {
				var attachIdeCommand AttachIdeCommand

				json.Unmarshal([]byte(msg), &attachIdeCommand)

				connectedBoard.detach()
				notify("attachIde", "")
				go monitorSerialPorts(attachIdeCommand.Arguments.Devices)
			} else {
				connectedBoard.reset(false)
				notify("boardAttached", "")
			}

		case "detachIde":
			StopMonitor = true
			connectedBoard.detach()

		case "boardReset":
			if connectedBoard != nil {
				notify("boardDetached", "")
				connectedBoard.consoleOut = false
				if connectedBoard.reset(false) {
					notify("boardReset", "")
					notify("boardAttached", "")
					connectedBoard.consoleOut = true
				}
			}

		case "boardStop":
			if connectedBoard != nil {
				notify("boardDetached", "")
				connectedBoard.consoleOut = false
				if connectedBoard.reset(false) {
					notify("boardReset", "")
					notify("boardAttached", "")
					connectedBoard.consoleOut = true
				}
			}

		case "boardGetDirContent":
			if connectedBoard != nil {
				var fsCommand CommandFileSystem

				json.Unmarshal([]byte(msg), &fsCommand)

				connectedBoard.consoleOut = false
				notify("boardGetDirContent", connectedBoard.getDirContent(fsCommand.Arguments.Path))
				connectedBoard.consoleOut = true
			}

		case "boardReadFile":
			if connectedBoard != nil {
				var fsCommand CommandFileSystem

				json.Unmarshal([]byte(msg), &fsCommand)

				connectedBoard.consoleOut = false
				notify("boardReadFile", base64.StdEncoding.EncodeToString(connectedBoard.readFile(fsCommand.Arguments.Path)))
				connectedBoard.consoleOut = true
			}

		case "boardWriteFile":
			if connectedBoard != nil {
				var fsCommand CommandFileSystem

				json.Unmarshal([]byte(msg), &fsCommand)

				content, err := base64.StdEncoding.DecodeString(fsCommand.Arguments.Content)
				if err == nil {
					connectedBoard.consoleOut = false
					connectedBoard.writeFile(fsCommand.Arguments.Path, content)
					notify("boardWriteFile", "")
					connectedBoard.consoleOut = true
				}
			}

		case "boardRunProgram":
			if connectedBoard != nil {
				var runCommand CommandRunProgram

				json.Unmarshal([]byte(msg), &runCommand)

				code, err := base64.StdEncoding.DecodeString(runCommand.Arguments.Code)
				if err == nil {
					connectedBoard.consoleOut = false
					connectedBoard.runProgram(runCommand.Arguments.Path, []byte(code))
					notify("boardRunProgram", "")
					connectedBoard.consoleOut = true
				}
			}

		case "boardRunCommand":
			if connectedBoard != nil {
				var runCommand CommandRunCommand

				json.Unmarshal([]byte(msg), &runCommand)

				code, err := base64.StdEncoding.DecodeString(runCommand.Arguments.Code)
				if err == nil {
					connectedBoard.consoleOut = false
					connectedBoard.runCode(code)
					response := connectedBoard.runCommand([]byte("_code()"))
					notify("boardRunCommand", base64.StdEncoding.EncodeToString([]byte(response)))
					connectedBoard.consoleOut = true
				}
			}

		case "boardUpgrade":
			if connectedBoard != nil {
				connectedBoard.consoleOut = false
				connectedBoard.upgrade()
				notify("boardUpgraded", "")
				connectedBoard.consoleOut = true
			}
		}
	}
}

func webSocketStart(exitChan chan int) {
	//generateCertificates()

	http.Handle("/", websocket.Handler(handler))

	go func() {
		log.Println("Starting non secure websocket server ...")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	}()

	//go func() {
	//log.Println("Starting secure websocket server ...")
	//if err := http.ListenAndServeTLS(":8081", "cert.pem", "key.pem", nil); err != nil {
	//	log.Fatal("ListenAndServe:", err)
	//}
	//}()
}
