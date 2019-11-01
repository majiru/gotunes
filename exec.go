package main

import (
	"log"
	"os/exec"
	"strings"
)

func dup(q []string) []string {
	s := make([]string, len(q))
	copy(s, q)
	return s
}

var dupmap map[string]struct{} = make(map[string]struct{})

func DJ(song chan string, ctl chan string, out chan []string, noDup bool) {
	var queue []string = nil
	cmd, done := play(<-song)

	for {
		select {
		case in := <-song:
			if noDup {
				if _, ok := dupmap[in]; ok {
					continue
				}
				if !strings.HasPrefix(in, "https://") {
					continue
				}
				dupmap[in] = struct{}{}
			}
			if cmd == nil {
				cmd, done = play(in)
			} else {
				queue = append(queue, in)
			}
		case err := <-done:
			if err != nil {
				log.Println(err)
			}
			if len(queue) == 0 {
				cmd = nil
				continue
			}
			s := queue[0]
			cmd, done = play(s)
			switch len(queue) {
			case 1:
				queue = nil
			default:
				queue = queue[1:]
			}
		case msg := <-ctl:
			switch msg {
			case "skip":
				if cmd != nil {
					cmd.Process.Kill()
					cmd = nil
				}
			}
		case out <- dup(queue):
		}
	}
}



func play(song string) (*exec.Cmd, chan error) {
	done := make(chan error)
	cmd := exec.Command("mpv", "--length=5:00", "--no-video", "--no-terminal", song)
	if err := cmd.Start(); err != nil {
		log.Println(err)
		return nil, nil
	}
	go func(d chan error) {
		done <- cmd.Wait()
	}(done)
	return cmd, done
}
