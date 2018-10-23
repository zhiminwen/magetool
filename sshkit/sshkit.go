package sshkit

import (
	"sync"
)

func Execute(targets []*SSHClient, cmd string) {
	var wg sync.WaitGroup
	for _, target := range targets {
		wg.Add(1)
		go func(target *SSHClient) {
			err := target.Execute(cmd)
			if err != nil {
				// error already printout in the Execute()
				// log.Printf("Error to execute for target: (%s:%s). error:%v", target.Host, target.Port, err)
			}
			wg.Done()
		}(target)
	}

	wg.Wait()
}

func ExecuteFunc(targets []*SSHClient, Fn func(*SSHClient)) {
	var wg sync.WaitGroup
	for _, target := range targets {
		wg.Add(1)
		go func(target *SSHClient) {
			Fn(target)
			wg.Done()
		}(target)
	}

	wg.Wait()
}

func Capture(target *SSHClient, cmd string) (string, error) {
	return target.Capture(cmd)
}
