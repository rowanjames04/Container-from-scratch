package main

import (
	"fmt"
	"io/ioutil"
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
	must(syscall.Mount("proc", "proc", "proc", 0, ""))
	must(syscall.Mount("thing", "mytemp", "tmpfs", 0, ""))
	
	cg()

	must(cmd.Run())

	must(syscall.Unmount("proc", 0))
	must(syscall.Unmount("thing", 0))

}

func cg() {
	cgroups := "sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	err := os.Mkdir(filepath.Join(pids, "r"), 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	must(ioutil.WriteFile(filepath.Join(pids, "r/pids.max"), []byte("20"), 0700))
	must(ioutil.WriteFile(filepath.Join(pids, "r/notify_on_release"), []byte("1"), 0700))
	must(ioutil.WriteFile(filepath.Join(pids, "r/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))

}

func must(err error) {
	if err != nil {
		panic(err)
	}
}