package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

var force = flag.Bool("force", false, "Bypass environment checks and run anyway")

func sessionsExist() (bool, error) {
	matches, err := filepath.Glob("/usr/share/xsessions/*.desktop")
	if err != nil {
		return false, err
	}
	return len(matches) > 0, nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getUsername() (string, error) {
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser != "" {
		return sudoUser, nil
	}
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.Username, nil
}

func getCurrentSessionID() (string, error) {
	out, err := exec.Command("loginctl", "list-sessions", "--no-legend").Output()
	if err != nil {
		return "", err
	}

	outputStr := string(out)
	lines := strings.Split(outputStr, "\n")

	username, err := getUsername()
	if err != nil {
		return "", err
	}

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			fmt.Printf("[DEBUG] Checking session line: %v\n", fields)
			if fields[2] == username {
				fmt.Printf("[DEBUG] Match found! Session ID: %s\n", fields[0])
				return fields[0], nil
			}
		}
	}

	return "", fmt.Errorf("session ID not found")
}

func isXorgAvailable() bool {
	_, err := exec.LookPath("Xorg")
	return err == nil
}

func isXorgInUse(sessionID string) bool {
	out, err := exec.Command("loginctl", "show-session", sessionID, "-p", "Type").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "Type=x11")
}

func main() {
	flag.Parse()
	if !*force {
		if !isXorgAvailable() {
			fmt.Println("WARNING: Xorg is not installed on this system. If you are sure you are running under Xorg, use --force to skip this check.")
			os.Exit(1)
		}
		fmt.Println("XORG available in the system.")
		sessionID, err := getCurrentSessionID()
		if err != nil {
			fmt.Println("WARNING: Could not determine your session ID. If you are sure you are running under Xorg, use --force to skip this check.")
			os.Exit(1)
		}
		fmt.Printf("INFO: Session ID found: %s\n", sessionID)
		if !isXorgInUse(sessionID) {
			fmt.Println("WARNING: Current session is not Xorg (likely Wayland). If you are sure you are running under Xorg, use --force to skip this check.")
			os.Exit(1)
		}
		fmt.Println("INFO: Xorg session confirmed.")
	} else {
		fmt.Println("[WARNING] Force mode may break your system. Press Enter to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Can't get home directory:", err)
		return
	}

	dconfUser := filepath.Join(homeDir, ".config", "dconf", "user")
	dconfBackup := filepath.Join(homeDir, ".config", "dconf", "user.bak")

	exist, err := sessionsExist()
	if err != nil {
		fmt.Println("Error checking sessions:", err)
		return
	}

	if exist {
		fmt.Println("Session files found, no need to reinstall.")
	} else {
		fmt.Println("No session files found, starting reinstall.")

		err = copyFile(dconfUser, dconfBackup)
		if err != nil {
			fmt.Println("Backup of dconf failed:", err)
		} else {
			fmt.Println("Backup of dconf done.")
		}

		fmt.Println("Reinstalling xserver-xorg-core, xorg, x11-common...")
		err = runCmd("sudo", "apt-get", "install", "--reinstall", "-y", "xserver-xorg-core", "xorg", "x11-common")
		if err != nil {
			fmt.Println("Failed reinstalling Xorg:", err)
			return
		}

		fmt.Println("Reinstalling gnome-session...")
		err = runCmd("sudo", "apt-get", "install", "--reinstall", "-y", "gnome-session")
		if err != nil {
			fmt.Println("Failed reinstalling gnome-session:", err)
			return
		}

		err = copyFile(dconfBackup, dconfUser)
		if err != nil {
			fmt.Println("Failed restoring dconf backup:", err)
		} else {
			fmt.Println("Restored dconf backup.")
		}
	}

	fmt.Println("Restarting gdm...")
	err = runCmd("sudo", "systemctl", "restart", "gdm")
	if err != nil {
		fmt.Println("Failed restarting gdm:", err)
		return
	}

	fmt.Println("All done.")
}
