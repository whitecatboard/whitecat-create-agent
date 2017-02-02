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

Available commands:

{"command": "attachIde", "arguments": "{}"}
{"command": "detachIde", "arguments": "{}"}

{"command": "boardInfo", "arguments": "{}"}
{"command": "boardReset, "arguments": "{}"}
{"command": "boardStop, "arguments": "{}"}
{"command": "boardGetDirContent", "arguments": {"path": "xxxx"}}
{"command": "boardReadFile", "arguments": {"path": "xxxx"}}
{"command": "boardRunCode", "arguments": {"path": "xxxx", "code": "xxxx"}}

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

type CommandRun struct {
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
		info = "{\"info\": " + connectedBoard.info + "}"

	case "boardRuntimeError":
		info = "{" + data + "}"

	case "boardGetDirContent":
		info = data

	case "boardReadFile":
		info = "{\"content\": \"" + data + "\"}"
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
			if connectedBoard == nil {
				var attachIdeCommand AttachIdeCommand

				json.Unmarshal([]byte(msg), &attachIdeCommand)

				connectedBoard.detach()
				notify("attachIde", "")
				go monitorSerialPorts(attachIdeCommand.Arguments.Devices)
			} else {
				notify("attachIde", "")
				notify("boardAttached", "")
			}

		case "boardReset":
			if connectedBoard != nil {
				notify("boardDetached", "")
				if connectedBoard.reset() {
					notify("boardReset", "")
					notify("boardAttached", "")
				}
			}

		case "boardStop":
			if connectedBoard != nil {
				notify("boardDetached", "")
				if connectedBoard.reset() {
					notify("boardReset", "")
					notify("boardAttached", "")
				}
			}

		case "boardGetDirContent":
			if connectedBoard != nil {
				var fsCommand CommandFileSystem

				json.Unmarshal([]byte(msg), &fsCommand)

				notify("boardGetDirContent", connectedBoard.getDirContent(fsCommand.Arguments.Path))
			}

		case "boardReadFile":
			if connectedBoard != nil {
				var fsCommand CommandFileSystem

				json.Unmarshal([]byte(msg), &fsCommand)

				notify("boardReadFile", base64.StdEncoding.EncodeToString(connectedBoard.readFile(fsCommand.Arguments.Path)))
			}

		case "boardWriteFile":
			if connectedBoard != nil {
				var fsCommand CommandFileSystem

				json.Unmarshal([]byte(msg), &fsCommand)

				content, err := base64.StdEncoding.DecodeString(fsCommand.Arguments.Content)
				if err == nil {
					connectedBoard.writeFile(fsCommand.Arguments.Path, content)
					notify("boardWriteFile", "")
				}
			}

		case "boardRunCode":
			if connectedBoard != nil {
				var runCommand CommandRun

				json.Unmarshal([]byte(msg), &runCommand)

				code, err := base64.StdEncoding.DecodeString(runCommand.Arguments.Code)
				if err == nil {
					connectedBoard.runCode(runCommand.Arguments.Path, []byte(code))
					notify("boardRunCode", "")
				}
			}

		}
	}
}

func webSocketStart(exitChan chan int) {
	log.Println("starting websocket server ...")

	http.Handle("/", websocket.Handler(handler))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
