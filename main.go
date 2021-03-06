package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strings"
	"syscall"
	"time"
)

// Get GoSh version information.
func version() {
	fmt.Println("\x1b[32;1mGoSh v1.0 - A terminal emulator written in Go!\x1b[0m")
	fmt.Println("\x1b[32;1mWritten by Munsy.\x1b[0m")
	fmt.Println("\x1b[32;1mLast updated: 12/22/2015\x1b[0m")
}

// Print the command prompt with relevant info.
func printCommandLine() {
	user, err := user.Current()
	if nil != err {
		panic("Error getting user information!")
	}

	hostname, err := os.Hostname()
	if nil != err {
		panic("Error getting hostname!")
	}

	cwd, err := os.Getwd()
	if nil != err {
		panic("Error getting current working directory!")
	}

	fmt.Printf("\x1b[34;1m%s\x1b[0m@\x1b[35;1m%s\x1b[0m \x1b[36;1m%s %% \x1b[0m", user.Username, hostname, cwd)
}

// Print working directory.
func pwd() {
	cwd, err := os.Getwd()
	if nil != err {
		fmt.Println("Error getting current working directory!")
		return
	}
	fmt.Println(cwd)
}

// User information lookup.
func finger(args []string) {
	var usr *user.User
	var err error

	if 1 == len(args) {
		usr, err = user.Current()
		if nil != err {
			fmt.Printf("\x1b[31;1mfinger: failed to get current user.\x1b[0m\n")
			return
		}
	} else {
		usr, err = user.Lookup(args[1])
		if nil != err {
			fmt.Printf("\x1b[31;1mfinger: %s: no such user.\x1b[0m\n", args[1])
			return
		}
	}

	fmt.Printf("\x1b[32;1mLogin:\x1b[0m %s \t%-25s\x1b[32;1mName:\x1b[0m %s \n", usr.Username, "", usr.Name)
	fmt.Printf("\x1b[32;1mUID:\x1b[0m %s \t%-25s\x1b[32;1mGID:\x1b[0m %s \n", usr.Uid, "", usr.Gid)
	fmt.Printf("\x1b[32;1mHome:\x1b[0m %s\n", usr.HomeDir)
}

// Prints environment variables.
func env() {
	env := os.Environ()
	for i := 0; i < len(env); i++ {
		fmt.Println(env[i])
	}
}

// Changes the current working directory.
func changeDir(args []string) {
	if 1 == len(args) {
		usr, err := user.Current()
		if nil != err {
			fmt.Printf("\x1b[31;1mfinger: failed to get current user.\x1b[0m\n")
		}
		if nil != os.Chdir(usr.HomeDir) {
			fmt.Printf("\x1b[31;1mcd: %s: No such file or directory\x1b[0m\n", usr.HomeDir)
		}
	} else if nil != os.Chdir(args[1]) {
		fmt.Printf("\x1b[31;1mcd: %s: No such file or directory\x1b[0m\n", args[1])
	}
}

// List directory contents
func ll(args []string) {
	var name string
	var err error

	path, err := os.Getwd()
	if err != nil {
		fmt.Printf("\x1b[31;1mls: Can't get current working directory\x1b[0m\n")
		return
	}

	if 1 == len(args) {
		name = path + "/."
	} else {
		if strings.HasPrefix(args[1], "/") {
			name = args[1]
		} else {
			name = path + "/" + args[1]
		}
	}

	file, err := os.Lstat(name)
	if nil != err {
		fmt.Printf("\x1b[31;1mls: Can't stat %s\x1b[0m\n", name)
		return
	}

	mode := file.Mode()

	if mode.IsDir() {
		llDir(name)
	} else {
		llFile(name)
	}
}

// Prints information about a directory.
func llDir(name string) {
	llFile(name + "/.")
	llFile(name + "/..")

	dir, err := os.Open(name)

	if nil != err {
		fmt.Printf("\x1b[31;1mls: Can't open %s\x1b[0m\n", name)
		return
	}

	dirs, err := dir.Readdirnames(0)

	if nil != err {
		fmt.Printf("\x1b[31;1mls: Can't read entries in %s\x1b[0m\n", name)
		return
	}

	for i := 0; i < len(dirs); i++ {
		llFile(name + "/" + dirs[i])
	}
}

// Prints information about a file.
func llFile(name string) {
	file, err := os.Lstat(name)
	if nil != err {
		fmt.Printf("\x1b[31;1mls: Can't stat %s\x1b[0m\n", name)
		return
	}

	perms := file.Mode().String()

	if strings.HasPrefix(perms, "L") {
		perms = perms[1:len(perms)]
		perms = "l" + perms
	}

	fmt.Printf("%s ", perms)

	var stNumLinks uint64
	sys := file.Sys()
	if sys != nil {
		if stat, ok := sys.(*syscall.Stat_t); ok {
			stNumLinks = uint64(stat.Nlink)
		}
	}

	fmt.Printf("%d\t", stNumLinks)
	fmt.Printf("%d\t", sys.(*syscall.Stat_t).Uid)
	fmt.Printf("%d\t", sys.(*syscall.Stat_t).Gid)
	fmt.Printf("%d\t", file.Size())
	fmt.Printf("%s ", file.ModTime().Format(time.UnixDate))
	fmt.Printf("%s", file.Name())

	if 0 != (file.Mode() & os.ModeSymlink) {
		linkname, err := os.Readlink(file.Name())
		if err != nil {
			fmt.Printf("\n")
			return
		}
		fmt.Printf(" -> %s\n", linkname)
	} else {
		fmt.Printf("\n")
	}
}

// List directory contents in compact form.
func ls(args []string) {
	var name string
	var err error

	path, err := os.Getwd()
	if err != nil {
		fmt.Printf("\x1b[31;1mls: Can't get current working directory\x1b[0m\n")
		return
	}

	if 1 == len(args) {
		name = path + "/."
	} else {
		if strings.HasPrefix(args[1], "/") {
			name = args[1]
		} else {
			name = path + "/" + args[1]
		}
	}

	file, err := os.Lstat(name)
	if nil != err {
		fmt.Printf("\x1b[31;1mls: Can't stat %s\x1b[0m\n", name)
		return
	}

	mode := file.Mode()

	if mode.IsDir() {
		dir, err := os.Open(name)
		if nil != err {
			fmt.Printf("\x1b[31;1mls: Can't open %s\x1b[0m\n", name)
			return
		}

		dirs, err := dir.Readdirnames(0)
		if nil != err {
			fmt.Printf("\x1b[31;1mls: Can't read entries in %s\x1b[0m\n", name)
			return
		}

		for i := 0; i < len(dirs); i++ {
			if strings.HasPrefix(dirs[i], ".") {
				continue
			}
			fmt.Printf("%s ", dirs[i])
		}
		if 0 != len(dirs) {
			fmt.Printf("\n")
		}
	} else {
		fmt.Printf("%s\n", name)
	}
}

func printOut(line string) {
	fmt.Printf("%s", line)
}

// Parse the string the user enters in to the command prompt.
func parseCommand(line string) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)

	go func() {
		for {
			<-s
		}
	}()
	cmd := exec.Command("/bin/sh", "-c", line)
	cmd.Stdout = os.Stdout //&out
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func main() {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)

	go func() {
		for {
			<-s
			fmt.Printf("\n")
		}
	}()

	printCommandLine()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		scanner.Scan()
		line := scanner.Text()
		args := strings.Split(line, " ")

		if "" == args[0] || " " == args[0] {
			continue
		} else if "ll" == args[0] {
			ll(args)
		} else if "cd" == args[0] {
			changeDir(args)
		} else if "version" == args[0] {
			version()
		} else if "exit" == args[0] {
			os.Exit(0)
		} else {
			parseCommand(line)
		}
		printCommandLine()
	}
}
