# What's The Whitecat Create Agent?

The Whitecat Create Agent is a small piece of software that runs on your computer, and allows the communication beetween a Whitecat Board and The Whitecat IDE. This is needed because the IDE must perform many operations that needs to use your computer hardware, and the IDE is on the cloud!. The communication beetween the agent and the IDE is made using websockets.

# How to build?

## If you don't want to modify the Whitecat Create Agent sources ...

1. Go to your Go's workspace location

   For example:

   ```lua
   cd gows
   ```

1. Download and install

   ```lua
   go get github.com/whitecatboard/whitecat-create-agent
   ```
   
   You can find the Whitecat Create Agent at the bin folder. For execute:
   
   Linux / OSX:
   
   ```lua
   ./bin/whitecat-create-agent
   ```
   
   Windows:
   
   ```lua
   bin\whitecat-create-agent.exe
   ```
   
## If you want to modify the Whitecat Create Agent sources ...

1. Go to your Go's workspace location

   For example:

   ```lua
   cd gows
   ```

1. Download and install

   ```lua
   go get github.com/whitecatboard/whitecat-create-agent
   ```

1. Go to the project source root

   ```lua
   cd src/github.com/whitecatboard/whitecat-create-agent
   ```

1. Make the changes on source code

1. Build project

   ```lua
   go build
   ```
   
   For execute:
   
   Linux / OSX:
   
   ```lua
   ./whitecat-create-agent
   ```
   
   Windows:
   
   ```lua
   whitecat-create-agent.exe
   ```




