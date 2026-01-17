package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

// docker         run image <cmd> <params> <- docker command
// go run main.go run       <cmd> <params> <- equivalent go command

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("bad command")
	}
}


func run() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]... ) ...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr {
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}
	
	must(cmd.Run())
}

func child() {
	fmt.Printf("Running %v \n", os.Args[2:])

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(syscall.Sethostname([]byte("container")))
	must(syscall.Chroot("/containers/ubuntufs"))
	must(os.Chdir("/"))

	must(syscall.Mount("proc", "/proc", "proc", 0, ""))
	must(syscall.Mount("sysfs", "/sys", "sysfs", 0, ""))
	must(syscall.Mount("cgroup2", "/sys/fs/cgroup", "cgroup2", 0, ""))
	must(syscall.Mount("tmpfs", "/mytemp", "tmpfs", 0, ""))
	
	cg()

	must(cmd.Run())

	must(syscall.Unmount("/mytemp", 0))
	must(syscall.Unmount("/sys/fs/cgroup", 0))
	must(syscall.Unmount("/proc", 0))
	must(syscall.Unmount("/sys", 0))
	

}

func cg() {
	cgroupPath := "/sys/fs/cgroup/mycontainer"

	must(os.MkdirAll(cgroupPath, 0755))

	// Limit number of processes
	must(os.WriteFile(
		filepath.Join(cgroupPath, "pids.max"),
		[]byte("20"),
		0644,
	))

	// Move this process into the cgroup
	must(os.WriteFile(
		filepath.Join(cgroupPath, "cgroup.procs"),
		[]byte(strconv.Itoa(os.Getpid())),
		0644,
	))

}

func must(err error) {
	if err != nil {
		panic(err)
	}
}