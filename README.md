<img src="https://cdn.nexusfn.net/file/2023/06/Lightyearnew.png">

Open source Fortnite launcher written in Go

To use with your own Cobalt dll, replace url in main.go with your own.

## How to build: 

- Download Go from [here](https://go.dev/dl/go1.20.5.windows-amd64.msi)
- Download the source code
- Extract zip file
- Open build.bat

Now you have a file called lightyear.exe, this is the launcher. Have fun!

## How to use your own Cobalt dll

- Start the launcher
- Exit the launcher
- Open `redirect.json` in the same folder as the launcher
- Change `name` to your dll name
- Change `download` to the url of your dll

Now zip your .exe into a zip file with the redirect.json in the same folder as the launcher if you want to share it with others for use.
