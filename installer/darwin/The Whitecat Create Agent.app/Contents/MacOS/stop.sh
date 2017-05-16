#!/bin/sh

/bin/launchctl unload -w /Library/LaunchAgents/org.whitecatboard.the-whitecat-create-agent.plist

osascript -e 'quit app "The Whitecat Create Agent"'