package api

import (
	"context"
	"errors"
	"net"
	"time"

	"../action"
	"../container"
	log "github.com/sirupsen/logrus"
)

// LinuxSignals valid Linux signal table
// http://www.comptechdoc.org/os/linux/programming/linux_pgsignals.html
var LinuxSignals = map[string]int{
	"SIGHUP":    1,
	"SIGINT":    2,
	"SIGQUIT":   3,
	"SIGILL":    4,
	"SIGTRAP":   5,
	"SIGIOT":    6,
	"SIGBUS":    7,
	"SIGFPE":    8,
	"SIGKILL":   9,
	"SIGUSR1":   10,
	"SIGSEGV":   11,
	"SIGUSR2":   12,
	"SIGPIPE":   13,
	"SIGALRM":   14,
	"SIGTERM":   15,
	"SIGSTKFLT": 16,
	"SIGCHLD":   17,
	"SIGCONT":   18,
	"SIGSTOP":   19,
	"SIGTSTP":   20,
	"SIGTTIN":   21,
	"SIGTTOU":   22,
	"SIGURG":    23,
	"SIGXCPU":   24,
	"SIGXFSZ":   25,
	"SIGVTALRM": 26,
	"SIGPROF":   27,
	"SIGWINCH":  28,
	"SIGIO":     29,
	"SIGPWR":    30,
}

var (
	Client     container.Client
	Chaos      action.Chaos
	TopContext context.Context
)

func runChaosCommand(cmd interface{}, interval time.Duration, names []string, pattern string, chaosFn func(context.Context, container.Client, []string, string, interface{}) error) {
	// create Time channel for specified interval
	var tick <-chan time.Time
	if interval == 0 {
		tick = time.NewTimer(interval).C
	} else {
		tick = time.NewTicker(interval).C
	}

	// handle the 'chaos' command
	ctx, cancel := context.WithCancel(TopContext)
	for {
		// cancel current context on exit
		defer cancel()
		// run chaos function
		if err := chaosFn(ctx, Client, names, pattern, cmd); err != nil {
			log.Error(err)
		}
		// wait for next timer tick or cancel
		select {
		case <-TopContext.Done():
			return // not to leak the goroutine
		case <-tick:
			if interval == 0 {
				return // not to leak the goroutine
			}
			log.Debug("Next chaos execution (tick) ...")
		}
	}
}

/**
* Kill a set of containers identified by names []string or pattern string
 */
func Kill(signal string, interval time.Duration, names []string, pattern string) error {
	if _, ok := LinuxSignals[signal]; !ok {
		err := errors.New("Unexpected signal: " + signal)
		log.Error(err)
		return err
	}
	runChaosCommand(action.CommandKill{Signal: signal}, interval, names, pattern, Chaos.KillContainers)
	return nil
}

/**
* Add delay to incoming network packet for a set of containers identified by names []string or pattern string
 */
func NetemDelay(interval time.Duration, duration time.Duration, names []string, pattern string,
	netInterface string, ips []net.IP, image string, time int, jitter int, correlation float64, distribution string) error {
	// pepare netem delay command
	delayCmd := action.CommandNetemDelay{
		NetInterface: netInterface,
		IPs:          ips,
		Duration:     duration,
		Time:         time,
		Jitter:       jitter,
		Correlation:  correlation,
		Distribution: distribution,
		Image:        image,
	}
	runChaosCommand(delayCmd, interval, names, pattern, Chaos.NetemDelayContainers)
	return nil
}

func NetemLossRandom(interval time.Duration, duration time.Duration, names []string, pattern string,
	netInterface string, ips []net.IP, image string, correlation float64, percent float64) error {
	// pepare netem loss command
	delayCmd := action.CommandNetemLossRandom{
		NetInterface: netInterface,
		IPs:          ips,
		Duration:     duration,
		Percent:      percent,
		Correlation:  correlation,
		Image:        image,
	}
	runChaosCommand(delayCmd, interval, names, pattern, Chaos.NetemLossRandomContainers)
	return nil
}

func NetemLossRate(interval time.Duration, duration time.Duration, names []string, pattern string,
	netInterface string, ips []net.IP, image string,
	p13 float64, p31 float64, p32 float64, p23 float64, p14 float64) error {
	// pepare netem loss command
	delayCmd := action.CommandNetemLossState{
		NetInterface: netInterface,
		IPs:          ips,
		Duration:     duration,
		P13:          p13,
		P31:          p31,
		P32:          p32,
		P23:          p23,
		P14:          p14,
		Image:        image,
	}
	runChaosCommand(delayCmd, interval, names, pattern, Chaos.NetemLossStateContainers)
	return nil
}
