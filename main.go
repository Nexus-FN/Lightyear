package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"bufio"

	"github.com/fatih/color"
	"go.zoe.im/injgo"
)

type File struct {
	URL  string
	Name string
}

// Constants for DLL injection because i always forget what this does
const (
	PROCESS_ALL_ACCESS     = 0x1F0FFF
	MEM_COMMIT             = 0x1000
	MEM_RESERVE            = 0x2000
	PAGE_EXECUTE_READWRITE = 0x40
	STD_OUTPUT_HANDLE = -11
)

// Function prototypes for DLL injection
var (
	procVirtualAllocEx     = syscall.NewLazyDLL("kernel32.dll").NewProc("VirtualAllocEx")
	procWriteProcessMemory = syscall.NewLazyDLL("kernel32.dll").NewProc("WriteProcessMemory")
	procCreateRemoteThread = syscall.NewLazyDLL("kernel32.dll").NewProc("CreateRemoteThread")
)

func main() {

	if !fileExists("redirect.json") {
		file := []byte(`{ "name": "Buzz.dll", "download": "https://cdn.nexusfn.net/file/2023/06/TV.dll" }`)
		err := ioutil.WriteFile("redirect.json", file, 0644)
		if err != nil {
			panic(err)
		}
	}

	fileData, err := ioutil.ReadFile("redirect.json")
	if err != nil {
		panic(err)
	}

	var jsonData map[string]interface{}
	err = json.Unmarshal(fileData, &jsonData)
	if err != nil {
		panic(err)
	}

	dllName, ok := jsonData["name"].(string)
	if !ok {
		panic("Invalid JSON structure: 'name' key is missing or not a string")
	}

	dllDownload, ok := jsonData["download"].(string)
	if !ok {
		panic("Invalid JSON structure: 'download' key is missing or not a string")
	}

	if !strings.HasSuffix(dllName, ".dll") {
		dllName += ".dll"
	}

	fileList := []File{
		{URL: dllDownload, Name: dllName},
		{URL: "https://cdn.discordapp.com/attachments/958139296936783892/1000707724507623424/FortniteClient-Win64-Shipping_BE.exe", Name: "FortniteClient-Win64-Shipping_BE.exe"},
		{URL: "https://cdn.discordapp.com/attachments/958139296936783892/1000707724818006046/FortniteLauncher.exe", Name: "FortniteLauncher.exe"},
	}

	localappdata := fmt.Sprintf("%s/AppData/Local/Lightyear", os.Getenv("USERPROFILE"))

	color.Magenta(`
	╦  ┬┌─┐┬ ┬┌┬┐┬ ┬┌─┐┌─┐┬─┐
	║  ││ ┬├─┤ │ └┬┘├┤ ├─┤├┬┘
	╩═╝┴└─┘┴ ┴ ┴  ┴ └─┘┴ ┴┴└─
	`)

	if !folderExists(localappdata) {
		createFolder(localappdata)
	}

	for _, file := range fileList {
		filename := filepath.Base(file.Name)
		localPath := filepath.Join(localappdata, filename)

		if !fileExists(localPath) {
			color.White("Downloading missing file %s", filename)
			err := downloadFile(file.URL, localPath)
			if err != nil {
				panic(err)
			}
		}
	}

	//Ask for input
	var input string
	color.White(`Select your option:
	1. Start Fortnite
	2. Change Fortnite path
	3. Change email and password
	`)

	fmt.Scanln(&input)

	switch input {
	case "1":
		go runFortnite(localappdata, dllName)
		var inout string
		color.White("Press enter to exit")
		fmt.Scanln(&inout)
	case "2":
		changePath(localappdata)
		main()
	case "3":
		color.White("Please enter your email")
		var email string
		fmt.Scanln(&email)
		color.White("Please enter your password")
		var password string
		fmt.Scanln(&password)

		emailFile, err := os.Create(localappdata + "/email.txt")
		if err != nil {
			panic(err)
		}
		defer emailFile.Close()

		fmt.Fprintf(emailFile, "%s", email)

		passwordFile, err := os.Create(localappdata + "/password.txt")
		if err != nil {
			panic(err)
		}
		defer passwordFile.Close()
		fmt.Fprintf(passwordFile, "%s", password)

		password = strings.Repeat("*", len(password))
	}

}

func runFortnite(localappdata string, dllName string) {

	file, err := os.Open(localappdata + "/path.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		path := scanner.Text()

		if !folderExists(path + `/FortniteGame`) {
			color.Red("Invalid path, please try again")
			main()
			return
		}

		password, err := readFile(localappdata + "/password.txt")
		if err != nil {
			panic(err)
		}

		email, err := readFile(localappdata + "/email.txt")
		if err != nil {
			panic(err)
		}

		args := []string{
			"-log",
			"-epicapp=Fortnite",
			"-epicenv=Prod",
			"-epiclocale=en-us",
			"-epicportal",
			"-skippatchcheck",
			"-nobe",
			"-fromfl=eac",
			"-fltoken=3db3ba5dcbd2e16703f3978d",
			"-caldera=eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjoiYmU5ZGE1YzJmYmVhNDQwN2IyZjQwZWJhYWQ4NTlhZDQiLCJnZW5lcmF0ZWQiOjE2Mzg3MTcyNzgsImNhbGRlcmFHdWlkIjoiMzgxMGI4NjMtMmE2NS00NDU3LTliNTgtNGRhYjNiNDgyYTg2IiwiYWNQcm92aWRlciI6IkVhc3lBbnRpQ2hlYXQiLCJub3RlcyI6IiIsImZhbGxiYWNrIjpmYWxzZX0.VAWQB67RTxhiWOxx7DBjnzDnXyyEnX7OljJm-j2d88G_WgwQ9wrE6lwMEHZHjBd1ISJdUO1UVUqkfLdU5nofBQ",
			fmt.Sprintf("-AUTH_LOGIN=%s", email),
			fmt.Sprintf("-AUTH_PASSWORD=%s", password),
			"-AUTH_TYPE=epic",
		}

		color.White("Starting Fortnite...")

		startLauncher(localappdata + "/FortniteLauncher.exe")
		startLauncher(localappdata + "/FortniteClient-Win64-Shipping_BE.exe")
		startShipping(path, args)

		process, err := injgo.FindProcessByName("FortniteClient-Win64-Shipping.exe")
		if err != nil {
			panic(err)
		} else {
			err := injectDll(uint32(process.ProcessID), filepath.Join(localappdata, dllName))
			if err != nil {
				panic(err)
			}
			println("Injected")
		}

	}

}

func injectDll(processID uint32, dllPath string) error {
	// Open the target process with PROCESS_ALL_ACCESS
	hProcess, err := syscall.OpenProcess(PROCESS_ALL_ACCESS, false, processID)
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(hProcess)

	// Allocate memory in the target process to store the DLL path
	dllPathAddr, _, err := procVirtualAllocEx.Call(
		uintptr(hProcess),
		0,
		uintptr(len(dllPath)),
		MEM_RESERVE|MEM_COMMIT,
		PAGE_EXECUTE_READWRITE,
	)
	if dllPathAddr == 0 {
		return err
	}

	// Write the DLL path into the target process's memory
	dllPathBytes := []byte(dllPath)
	var bytesWritten uintptr
	_, _, err = procWriteProcessMemory.Call(
		uintptr(hProcess),
		dllPathAddr,
		uintptr(unsafe.Pointer(&dllPathBytes[0])),
		uintptr(len(dllPathBytes)),
		uintptr(unsafe.Pointer(&bytesWritten)),
	)
	if bytesWritten != uintptr(len(dllPathBytes)) {
		return err
	}

	// Load the kernel32.dll module in the target process
	kernel32, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		return err
	}
	defer syscall.FreeLibrary(kernel32)

	// Obtain the address of the LoadLibraryA function
	loadLibraryAddr, err := syscall.GetProcAddress(kernel32, "LoadLibraryA")
	if err != nil {
		return err
	}

	// Create a remote thread in the target process to execute the LoadLibraryA function
	hThread, _, err := procCreateRemoteThread.Call(
		uintptr(hProcess),
		0,
		0,
		loadLibraryAddr,
		dllPathAddr,
		0,
		0,
	)
	if hThread == 0 {
		return err
	}
	defer syscall.CloseHandle(syscall.Handle(hThread))

	return nil
}

func startLauncher(path string) {
	println("Starting launcher")
	if !fileExists(path) {
		color.Red("Launcher not found, please try again")
		return
	}
	cmd := exec.Command(path)
	cmd.Start()
}

func startShipping(gamePath string, args []string) {
	println("Starting shipping")
	if !fileExists(filepath.Join(gamePath, "FortniteGame", "Binaries", "Win64", "FortniteClient-Win64-Shipping.exe")) {
		color.Red("Shipping not found, please try again")
		return
	}
	cmd := exec.Command(filepath.Join(gamePath, "FortniteGame", "Binaries", "Win64", "FortniteClient-Win64-Shipping.exe"))
	cmd.Args = append(cmd.Args, args...)
	cmd.Start()
}

func changePath(localappdata string) {
	fmt.Println("Please enter your Fortnite path")
	reader := bufio.NewReader(os.Stdin)
	path, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	path = strings.TrimSpace(path)

	if !folderExists(path + `/FortniteGame`) {
		color.Red("Invalid path, please try again")
		color.White(path)
		changePath(localappdata)
	}

	file, err := os.Create(localappdata + "/path.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fmt.Fprintf(file, "%s", path)

	clearConsole()
}

func readFile(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	content := string(data)
	return content, nil
}

func folderExists(folder string) bool {
	_, err := os.Stat(folder)
	return err == nil
}

func createFolder(folder string) {
	err := os.MkdirAll(folder, 0755)
	if err != nil {
		panic(err)
	}
}

func fileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

func downloadFile(url string, outputPath string) error {
	// Create the output file
	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Perform the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if the response status is successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file, status code: %d", resp.StatusCode)
	}

	// Copy the response body to the output file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func clearConsole() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "linux", "darwin":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		fmt.Println("Unable to clear console. Unsupported operating system.")
	}
}