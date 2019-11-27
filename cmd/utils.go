package cmd

import (
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)
func isDebug() bool {
	return false
}
func isTrace() bool {
	return false
}

func ValidateCommand(exitCode int){
  if exitCode >0 {
  	log.Panic("Invalid Exit Code:" + strconv.Itoa(exitCode))
  }
}
func Command(name string, arg...string) int{
	if isDebug() {
		//log.Println(name,arg)
		return 0
	}else{
		cmd := exec.Command(name, arg...)
		var waitStatus syscall.WaitStatus
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if isTrace() {
			log.Println("Exec: " + strings.Join(cmd.Args, " "))
		}
		if err :=cmd.Run(); err!=nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				waitStatus = exitError.Sys().(syscall.WaitStatus)
				return waitStatus.ExitStatus()
			}else if execError, ok := err.(*exec.Error); ok{
				if isTrace(){
					log.Println(execError.Error())
				}
				return 999
			}else{
				log.Panic("Process error not controlled")
				return 998
			}
		}else{
			waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
			return waitStatus.ExitStatus()
		}
		log.Panic("Process error not controlled2")
		return 997
	}
}



func CommandCaptureOutput(name string, arg...string) (string,error){
	return CommandCaptureOutputStdin("",name,arg...)
}

func CommandCaptureOutputStdin(stdin string, name string, arg...string) (string, error) {
	if isDebug() {
		log.Println(name,arg)
		return "oQV6w3iAV7q3kAP6nYRfr54uFaCy7gnOZ/u1SLGUwCw=",nil
	}else {
		cmd := exec.Command(name, arg...)
		log.Println("Exec: " + strings.Join(cmd.Args, " "))
		var waitStatus syscall.WaitStatus
		cmd.Stderr = os.Stderr
		if stdin != "" {
			stdinpipe, err := cmd.StdinPipe()
			io.WriteString(stdinpipe, stdin)
			stdinpipe.Close()
			if err != nil {
				return "", err
			}
		}

		outputBytes, err := cmd.Output()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				waitStatus = exitError.Sys().(syscall.WaitStatus)
				if waitStatus.ExitStatus() != 0 {
					return "", errors.New("invalid exit code:" + strconv.Itoa(waitStatus.ExitStatus()))
				}

			} else if execError, ok := err.(*exec.Error); ok {
				return "", execError
			} else {
				log.Panic("Process error not controlled")
				return "", errors.New("Unknow process error")
			}
		} else {
			waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
			if waitStatus.ExitStatus() != 0 {
				return "", errors.New("invalid exit code:" + strconv.Itoa(waitStatus.ExitStatus()))
			}
		}

		return strings.TrimSpace(string(outputBytes)), nil
	}
}


