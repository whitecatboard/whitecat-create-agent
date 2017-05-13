go build -ldflags="-H windowsgui"
cp whitecat-create-agent.exe installer\windows\wccagent.exe
del whitecat-create-agent.exe
