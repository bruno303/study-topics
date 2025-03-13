package monitor

import "context"

type FsMonitor interface {
	Start(ctx context.Context)
}

type Command interface {
	Run(ctx context.Context)
	Stop(ctx context.Context)
}

type MonitorManager struct {
	fsMonitor FsMonitor
	command   Command
}

func NewMonitorManager(fsMonitor FsMonitor, command Command) MonitorManager {
	return MonitorManager{
		fsMonitor: fsMonitor,
		command:   command,
	}
}

func (m MonitorManager) Start(ctx context.Context) {
	go func() {
		m.command.Run(ctx)
	}()
	m.fsMonitor.Start(ctx)
}

func (m MonitorManager) Stop(ctx context.Context) {
	m.command.Stop(ctx)
}
