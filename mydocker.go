package main

import (
	"fmt"
	"myDocker/util"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

const (
	daemonPath        = "./config/net_init.sh"
	destroyPath       = ""
	netPairPath       = "./config/pair_init.sh"
	netJsonPath       = "./config/ip.json"
	containerInitPath = "./config/container_init.sh"
	cgroupInitPath    = "./config/cgroup_init.sh"
	CGROUPNAME        = "__demo__"
	echoPath          = "./config/echo.sh"
)

// a example instr: "go run mydocker.go run <cmd> <args>"
func main() {
	//os.Arg = [xxx/new run echo xxx]
	switch os.Args[1] {
	case "run":
		run()
	case "__container__":
		container() //this program shall never be called by user, so I altered the parameter
		//instead, the run() func will call the container()
	case "daemon":
		daemon()
	case "destroy":
		destroy()
	case "help":
		help()
	default:
		help()
	}
}

/*********************************       main implementations            ************************/
// this function will literally build up a container
func container() {
	fmt.Printf("[container]: container run procedure [%v] on process [%d] \n", os.Args[2:], os.Getpid()) // ip echo xxx
	containerIp := os.Args[2]

	//pair Init
	runCommand(containerInitPath,
		[]string{strconv.FormatInt(util.InetAtoN(containerIp), 10), containerIp}...)

	fmt.Println("[container]: initialization completed")

	cmd := exec.Command(os.Args[3], os.Args[4:]...)
	//fmt.Printf("container id:%d", cmd)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(syscall.Sethostname([]byte("container_hostname")))
	must(syscall.Chroot("./resource/busybox"))
	must(os.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))
	must(cmd.Run())
	must(syscall.Unmount("proc", 0))
	fmt.Println("[container]：exit successful!! that good!!")
}

//this function is for establishing the quasi-daemon process for a docker
func run() {
	fmt.Printf("[host]: host run procedure [%v] on process [%d] \n", os.Args[2:], os.Getpid()) // [echo xxx]

	newIp := initNet(netJsonPath)

	// ! here we begin to set up a new process to call the container()
	// cmd = [xxx/new container echo xxx]
	cmd := exec.Command("/proc/self/exe", append([]string{"__container__", newIp}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
		Unshareflags: syscall.CLONE_NEWNS,
	}
	must(cmd.Start()) // Start() is a non-blocking call comparing to Run() is a blocking call

	// ! here we will set up the network links between the bridge and the node
	// pairInit
	containerPid := cmd.Process.Pid
	runCommand(netPairPath, []string{strconv.FormatInt(util.InetAtoN(newIp), 10),
		strconv.Itoa(containerPid)}...)

	restrictResource(containerPid, CGROUPNAME)

	must(cmd.Wait())

	// cleanups
	releaseNet(netJsonPath, newIp)
	fmt.Println("[host]: program exited")
}

// set up the environments required for the docker
func daemon() {
	runCommand(daemonPath)                 // 建网桥
	runCommand(cgroupInitPath, CGROUPNAME) // 建cgroup
}

//clean up the environments required for the docker
func destroy() {
	//runCommand(destroyPath)
	releaseResouce(CGROUPNAME)
}

// this func will print help infos for the user
func help() {
	helpInfo := `
	HELP INFORMATION
	
	DESCRIPTION:
				a toy-like self made demo docker 
	SYNOPSIS:
				go run mydocker.go [options][arguments]
	OPTIONS:
				daemon
					set up the environments required for the docker
				run
					build up a docker
				destroy
					clean up the environments set for the docker
				help
					print help information `
	fmt.Println(helpInfo)
}

/******************************     assistant functions       *********************/

func must(err error) {

	if err != nil {
		// exit status 130 代表程序为ctrl+c退出，不算异常。
		/**
		reference：https://stackoverflow.com/questions/29887088/java-program-exit-with-code-130
		*/
		if err.Error() == "exit status 130" {
			return
		} else {
			panic(err)
		}
	}
}

//
//  runCommand
//  @Description: et to run a command, for example: ./config/echo.sh hello
//  @param _cmd: command or script file
//  @param args: command args
//
func runCommand(_cmd string, args ...string) {

	cmd := exec.Command(_cmd, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	must(cmd.Run())
}

/*******************************    network related functions  *********************/
//
//  initNet
//  @Description: 该函数通过读取netJsonPath文件，来判断哪一个ip可以被使用，然后将该ip的状态置1，并写回文件。
//  @param netJsonPath netJsonPath文件地址
//  @return string 返回ip地址，例如10.0.0.19
//
func initNet(netJsonPath string) string {
	iPAllocation := util.NewIPAllocation(netJsonPath)
	newIp := util.AllocationIp(iPAllocation)
	if newIp != "" {
		iPAllocation.Ip[newIp] = 1
		util.WriteIPAllocationToFile(iPAllocation, netJsonPath)
		return newIp
	} else {
		panic("unable to allocate IP address")
	}
}

//
//  releaseNet
//  @Description: 当容器关闭时，需要释放ip，该函数读取netJsonPath，然后将对应的ip的状态置为0，代表该ip未被使用，并写回文件。
//  @param netJsonPath
//  @param releaseIp 需要释放的ip
//
func releaseNet(netJsonPath string, releaseIp string) {
	iPAllocation := util.NewIPAllocation(netJsonPath)
	iPAllocation.Ip[releaseIp] = 0
	util.WriteIPAllocationToFile(iPAllocation, netJsonPath)
}

/******************** resource control related functions  **********/
func restrictResource(pid int, cgroupName string) {
	//runCommand("echo", []string{strconv.Itoa(pid), ">>", "/sys/fs/cgroup/cpu/" + cgroupName + "/tasks"}...)
	runCommand(echoPath, []string{strconv.Itoa(pid), "/sys/fs/cgroup/cpu/" + cgroupName + "/tasks"}...)
}

func releaseResouce(cgroupName string) {
	runCommand("cgdelete", "cpu:"+cgroupName)
}
